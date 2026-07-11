package cacheflow

import "testing"

type fixedCountEngine int

func (e fixedCountEngine) EncodeOrdinary(string) []int { return nil }
func (e fixedCountEngine) Decode([]int) string         { return "" }
func (e fixedCountEngine) CountTokens(string) int      { return int(e) }

func TestAlignPrompt(t *testing.T) {
	tests := []struct {
		name  string
		count int
		block int
		want  Alignment
	}{
		{
			name:  "already aligned",
			count: 1024,
			block: 128,
			want: Alignment{
				CurrentTokens:     1024,
				BlockSize:         128,
				PreviousBlockSize: 1024,
				NextBlockSize:     1024,
				IsAligned:         true,
				IsEligible:        true,
			},
		},
		{
			name:  "near boundary",
			count: 990,
			block: 128,
			want: Alignment{
				CurrentTokens:     990,
				BlockSize:         128,
				PreviousBlockSize: 896,
				NextBlockSize:     1024,
				Remainder:         94,
				PaddingNeeded:     34,
				IsEligible:        true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAligner(fixedCountEngine(tt.count)).AlignPrompt("ignored", tt.block)
			assertAlignment(t, got, tt.want)
		})
	}
}

func TestAlignPromptProfileMinimum(t *testing.T) {
	report := NewAligner(fixedCountEngine(900)).AlignPromptToProfile("ignored", Profile{
		Name:          "test",
		BlockSize:     128,
		MinimumTokens: 1024,
	})

	want := Alignment{
		CurrentTokens:      900,
		BlockSize:          128,
		MinimumTokens:      1024,
		PreviousBlockSize:  896,
		NextBlockSize:      1024,
		Remainder:          4,
		PaddingNeeded:      124,
		TokensUntilMinimum: 124,
		IsEligible:         false,
	}
	assertAlignment(t, report, want)
}

func TestAlignInvalidConfig(t *testing.T) {
	if got := NewAligner(nil).AlignPrompt("x", 128); got.StrategyHint == "" {
		t.Fatal("nil engine returned empty strategy hint")
	}
	if got := NewAligner(fixedCountEngine(10)).AlignPrompt("x", 0); got.StrategyHint == "" {
		t.Fatal("invalid block returned empty strategy hint")
	}
}

func TestProfileOpenAI(t *testing.T) {
	report := NewAligner(fixedCountEngine(1025)).AlignPromptToProfile("ignored", ProfileOpenAI)
	if report.BlockSize != 128 {
		t.Fatalf("OpenAI block size = %d, want 128", report.BlockSize)
	}
	if report.MinimumTokens != 1024 {
		t.Fatalf("OpenAI minimum = %d, want 1024", report.MinimumTokens)
	}
	if report.PaddingNeeded != 127 || report.NextBlockSize != 1152 {
		t.Fatalf("OpenAI alignment = padding %d next %d", report.PaddingNeeded, report.NextBlockSize)
	}
}

func assertAlignment(t *testing.T, got Alignment, want Alignment) {
	t.Helper()
	if got.CurrentTokens != want.CurrentTokens ||
		got.BlockSize != want.BlockSize ||
		got.MinimumTokens != want.MinimumTokens ||
		got.PreviousBlockSize != want.PreviousBlockSize ||
		got.NextBlockSize != want.NextBlockSize ||
		got.Remainder != want.Remainder ||
		got.PaddingNeeded != want.PaddingNeeded ||
		got.TokensUntilMinimum != want.TokensUntilMinimum ||
		got.IsAligned != want.IsAligned ||
		got.IsEligible != want.IsEligible {
		t.Fatalf("Alignment = %+v, want %+v", got, want)
	}
	if got.StrategyHint == "" {
		t.Fatal("empty StrategyHint")
	}
}
