package cacheflow

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ron2111/omnitoken"
)

// TraceItem is one rendered prompt or structured prompt in a workload trace.
type TraceItem struct {
	ID     string       `json:"id,omitempty"`
	Model  string       `json:"model,omitempty"`
	Prompt string       `json:"prompt,omitempty"`
	Parts  []PromptPart `json:"parts,omitempty"`
}

// PromptPart describes a named prompt section. Stable is user-provided metadata
// used for diagnostics; simulation always analyzes the rendered token prefix.
type PromptPart struct {
	Name   string `json:"name,omitempty"`
	Stable bool   `json:"stable,omitempty"`
	Text   string `json:"text,omitempty"`
}

// SimulationOptions controls prompt-cache trace analysis.
type SimulationOptions struct {
	Profile        Profile `json:"profile"`
	DetectBreakers bool    `json:"detect_breakers"`
}

// SimulationReport summarizes token-level cache planning for a trace.
type SimulationReport struct {
	Profile               Profile        `json:"profile"`
	Items                 []ItemReport   `json:"items"`
	TotalTokens           int            `json:"total_tokens"`
	CommonPrefixTokens    int            `json:"common_prefix_tokens"`
	ReusablePrefixTokens  int            `json:"reusable_prefix_tokens"`
	ReusablePrefixPercent float64        `json:"reusable_prefix_percent"`
	CacheBreakerHints     []CacheBreaker `json:"cache_breaker_hints,omitempty"`
	Summary               string         `json:"summary"`
}

// ItemReport describes one trace item in a simulation report.
type ItemReport struct {
	ID                  string    `json:"id,omitempty"`
	Model               string    `json:"model,omitempty"`
	PromptTokens        int       `json:"prompt_tokens"`
	CommonPrefixTokens  int       `json:"common_prefix_tokens"`
	DynamicSuffixTokens int       `json:"dynamic_suffix_tokens"`
	Alignment           Alignment `json:"alignment"`
}

// CacheBreaker is a best-effort warning about dynamic data that can reduce cache hits.
type CacheBreaker struct {
	ItemID  string `json:"item_id,omitempty"`
	Part    string `json:"part,omitempty"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

// Simulate analyzes a batch of rendered prompts for stable token prefixes.
func Simulate(engine omnitoken.ModelEngine, items []TraceItem, opts SimulationOptions) SimulationReport {
	profile := opts.Profile
	if profile.BlockSize == 0 {
		profile = ProfileOpenAI
	}
	if engine == nil || len(items) == 0 {
		return SimulationReport{Profile: profile, Summary: "No trace items to analyze"}
	}

	tokensByItem := make([][]int, len(items))
	reports := make([]ItemReport, len(items))
	totalTokens := 0
	for i, item := range items {
		prompt := item.RenderedPrompt()
		tokens := engine.EncodeOrdinary(prompt)
		tokensByItem[i] = tokens
		totalTokens += len(tokens)
		reports[i] = ItemReport{
			ID:           item.ID,
			Model:        item.Model,
			PromptTokens: len(tokens),
			Alignment:    AlignTokenCount(len(tokens), profile),
		}
	}

	commonPrefix := commonTokenPrefix(tokensByItem)
	reusablePrefix := reusablePrefix(commonPrefix, profile)
	for i := range reports {
		reports[i].CommonPrefixTokens = commonPrefix
		reports[i].DynamicSuffixTokens = reports[i].PromptTokens - commonPrefix
		if reports[i].DynamicSuffixTokens < 0 {
			reports[i].DynamicSuffixTokens = 0
		}
	}

	percent := 0.0
	if totalTokens > 0 && len(items) > 1 {
		percent = float64(reusablePrefix*(len(items)-1)) / float64(totalTokens) * 100
	}
	report := SimulationReport{
		Profile:               profile,
		Items:                 reports,
		TotalTokens:           totalTokens,
		CommonPrefixTokens:    commonPrefix,
		ReusablePrefixTokens:  reusablePrefix,
		ReusablePrefixPercent: percent,
		Summary:               simulationSummary(commonPrefix, reusablePrefix, profile),
	}
	if opts.DetectBreakers {
		report.CacheBreakerHints = DetectBreakers(items)
	}
	return report
}

// RenderedPrompt returns Prompt when present, otherwise concatenates Parts.
func (i TraceItem) RenderedPrompt() string {
	if i.Prompt != "" || len(i.Parts) == 0 {
		return i.Prompt
	}
	var b strings.Builder
	for _, part := range i.Parts {
		b.WriteString(part.Text)
	}
	return b.String()
}

// ReadJSONL reads cacheflow trace items from newline-delimited JSON.
func ReadJSONL(r io.Reader) ([]TraceItem, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	var items []TraceItem
	line := 0
	for scanner.Scan() {
		line++
		raw := strings.TrimSpace(scanner.Text())
		raw = strings.TrimPrefix(raw, "\ufeff")
		if raw == "" {
			continue
		}
		var item TraceItem
		if err := json.Unmarshal([]byte(raw), &item); err != nil {
			return nil, fmt.Errorf("cacheflow JSONL line %d: %w", line, err)
		}
		items = append(items, item)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func commonTokenPrefix(tokensByItem [][]int) int {
	if len(tokensByItem) == 0 {
		return 0
	}
	limit := len(tokensByItem[0])
	for _, tokens := range tokensByItem[1:] {
		if len(tokens) < limit {
			limit = len(tokens)
		}
	}
	for i := 0; i < limit; i++ {
		want := tokensByItem[0][i]
		for _, tokens := range tokensByItem[1:] {
			if tokens[i] != want {
				return i
			}
		}
	}
	return limit
}

func reusablePrefix(tokens int, profile Profile) int {
	if profile.BlockSize <= 0 || tokens < profile.MinimumTokens {
		return 0
	}
	return tokens - tokens%profile.BlockSize
}

func simulationSummary(commonPrefix int, reusablePrefix int, profile Profile) string {
	if reusablePrefix > 0 {
		return fmt.Sprintf("%d common prefix tokens are reusable at %s cache boundaries", reusablePrefix, profile.Name)
	}
	if commonPrefix > 0 {
		return fmt.Sprintf("%d common prefix tokens found, below the configured reusable cache boundary", commonPrefix)
	}
	return "No common token prefix found across trace items"
}
