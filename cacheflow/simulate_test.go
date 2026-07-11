package cacheflow

import (
	"strings"
	"testing"
)

type wordEngine struct{}

func (wordEngine) EncodeOrdinary(text string) []int {
	fields := strings.Fields(text)
	ids := make([]int, len(fields))
	for i, field := range fields {
		ids[i] = wordID(field)
	}
	return ids
}

func (wordEngine) Decode([]int) string { return "" }
func (wordEngine) CountTokens(text string) int {
	return len(strings.Fields(text))
}

func TestSimulateCommonPrefix(t *testing.T) {
	items := []TraceItem{
		{ID: "a", Prompt: "stable prefix policy user one"},
		{ID: "b", Prompt: "stable prefix policy user two"},
	}
	report := Simulate(wordEngine{}, items, SimulationOptions{Profile: Profile{Name: "test", BlockSize: 2, MinimumTokens: 2}})
	if report.CommonPrefixTokens != 4 {
		t.Fatalf("CommonPrefixTokens = %d, want 4", report.CommonPrefixTokens)
	}
	if report.ReusablePrefixTokens != 4 {
		t.Fatalf("ReusablePrefixTokens = %d, want 4", report.ReusablePrefixTokens)
	}
	if len(report.Items) != 2 || report.Items[0].DynamicSuffixTokens != 1 {
		t.Fatalf("Items = %+v", report.Items)
	}
}

func TestReadJSONLAndBreakers(t *testing.T) {
	input := strings.NewReader(`{"id":"1","parts":[{"name":"meta","stable":false,"text":"timestamp 2026-07-11T10:00:00Z request 123e4567-e89b-12d3-a456-426614174000"},{"name":"system","stable":true,"text":"stable"}]}
{"id":"2","prompt":"stable prompt"}
`)
	items, err := ReadJSONL(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("items len = %d", len(items))
	}
	hints := DetectBreakers(items)
	if len(hints) < 2 {
		t.Fatalf("hints = %+v, want timestamp and uuid", hints)
	}
}

func TestStructuredPartsRender(t *testing.T) {
	item := TraceItem{Parts: []PromptPart{{Text: "hello "}, {Text: "world"}}}
	if got := item.RenderedPrompt(); got != "hello world" {
		t.Fatalf("RenderedPrompt = %q", got)
	}
}

func wordID(s string) int {
	h := 0
	for i := 0; i < len(s); i++ {
		h = h*31 + int(s[i])
	}
	if h < 0 {
		return -h
	}
	return h
}
