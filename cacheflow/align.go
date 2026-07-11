// Package cacheflow provides dependency-free prompt-cache planning and trace analysis.
package cacheflow

import (
	"fmt"

	"github.com/ron2111/omnitoken"
)

// Profile describes local token-boundary rules for cache planning.
type Profile struct {
	Name          string `json:"name"`
	BlockSize     int    `json:"block_size"`
	MinimumTokens int    `json:"minimum_tokens"`
}

// Common cache-planning profiles. Provider behavior can change; use these as
// local planning helpers, not billing guarantees.
var (
	ProfileGeneric = Profile{Name: "generic", BlockSize: 1024}
	ProfileOpenAI  = Profile{Name: "openai", BlockSize: 128, MinimumTokens: 1024}
)

// Alignment describes how close a prompt or prefix is to a cache block boundary.
type Alignment struct {
	CurrentTokens      int    `json:"current_tokens"`
	BlockSize          int    `json:"block_size"`
	MinimumTokens      int    `json:"minimum_tokens"`
	PreviousBlockSize  int    `json:"previous_block_size"`
	NextBlockSize      int    `json:"next_block_size"`
	Remainder          int    `json:"remainder"`
	PaddingNeeded      int    `json:"padding_needed"`
	TokensUntilMinimum int    `json:"tokens_until_minimum"`
	IsAligned          bool   `json:"is_aligned"`
	IsEligible         bool   `json:"is_eligible"`
	StrategyHint       string `json:"strategy_hint"`
}

// Aligner evaluates prompt lengths against provider cache block sizes.
type Aligner struct {
	engine omnitoken.ModelEngine
}

// NewAligner creates a prompt-cache alignment helper for an engine.
func NewAligner(engine omnitoken.ModelEngine) *Aligner {
	return &Aligner{engine: engine}
}

// AlignPrompt evaluates prompt length against a custom cache block size.
func (a *Aligner) AlignPrompt(text string, providerBlockSize int) Alignment {
	return a.AlignPromptToProfile(text, Profile{Name: "custom", BlockSize: providerBlockSize})
}

// AlignPromptToProfile evaluates prompt length against a cache-planning profile.
func (a *Aligner) AlignPromptToProfile(text string, profile Profile) Alignment {
	if a == nil || a.engine == nil {
		return Alignment{StrategyHint: "Invalid cache alignment configuration"}
	}
	return AlignTokenCount(a.engine.CountTokens(text), profile)
}

// AlignTokenCount evaluates an already-computed token count against a profile.
func AlignTokenCount(tokens int, profile Profile) Alignment {
	blockSize := profile.BlockSize
	if blockSize <= 0 {
		return Alignment{CurrentTokens: tokens, StrategyHint: "Invalid cache alignment configuration"}
	}
	minimum := profile.MinimumTokens
	if minimum < 0 {
		minimum = 0
	}

	remainder := tokens % blockSize
	previous := tokens - remainder
	padding := 0
	if remainder == 0 {
		previous = tokens
	} else {
		padding = blockSize - remainder
	}
	next := tokens + padding
	eligible := tokens >= minimum
	untilMinimum := 0
	if !eligible {
		untilMinimum = minimum - tokens
		if next < minimum {
			next = roundUp(minimum, blockSize)
			padding = next - tokens
		}
	}

	return Alignment{
		CurrentTokens:      tokens,
		BlockSize:          blockSize,
		MinimumTokens:      minimum,
		PreviousBlockSize:  previous,
		NextBlockSize:      next,
		Remainder:          remainder,
		PaddingNeeded:      padding,
		TokensUntilMinimum: untilMinimum,
		IsAligned:          remainder == 0,
		IsEligible:         eligible,
		StrategyHint:       strategyHint(tokens, minimum, padding, remainder),
	}
}

func roundUp(value int, blockSize int) int {
	if value <= 0 {
		return 0
	}
	remainder := value % blockSize
	if remainder == 0 {
		return value
	}
	return value + blockSize - remainder
}

func strategyHint(tokens int, minimum int, padding int, remainder int) string {
	if minimum > 0 && tokens < minimum {
		return fmt.Sprintf("Prompt is %d tokens below the configured cache minimum", minimum-tokens)
	}
	if padding == 0 && remainder == 0 {
		return "Prompt is aligned to the configured cache block boundary"
	}
	return fmt.Sprintf("Prompt is %d tokens from the next configured cache block boundary", padding)
}
