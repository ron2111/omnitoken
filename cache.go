package omnitoken

import "fmt"

// CacheProfile describes token-boundary rules for cache planning.
type CacheProfile struct {
	Name          string
	BlockSize     int
	MinimumTokens int
}

// Common cache-planning profiles. Provider cache behavior can change; use these
// as local planning helpers, not billing guarantees.
var (
	CacheProfileGeneric = CacheProfile{Name: "generic", BlockSize: 1024}
	CacheProfileOpenAI  = CacheProfile{Name: "openai", BlockSize: 128, MinimumTokens: 1024}
)

// CacheReport describes how close a prompt is to a cache block boundary.
type CacheReport struct {
	CurrentTokens      int
	BlockSize          int
	MinimumTokens      int
	PreviousBlockSize  int
	NextBlockSize      int
	Remainder          int
	PaddingNeeded      int
	TokensUntilMinimum int
	IsAligned          bool
	IsEligible         bool
	StrategyHint       string
}

// CacheAligner evaluates prompt lengths against provider cache block sizes.
type CacheAligner struct {
	engine ModelEngine
}

// NewCacheAligner creates a prompt cache alignment helper for an engine.
func NewCacheAligner(engine ModelEngine) *CacheAligner {
	return &CacheAligner{engine: engine}
}

// AlignPrompt evaluates prompt lengths to hit exact pricing or cache tier boundaries.
func (c *CacheAligner) AlignPrompt(text string, providerBlockSize int) CacheReport {
	return c.AlignPromptToProfile(text, CacheProfile{Name: "custom", BlockSize: providerBlockSize})
}

// AlignPromptToProfile evaluates prompt length against a cache-planning profile.
func (c *CacheAligner) AlignPromptToProfile(text string, profile CacheProfile) CacheReport {
	if c == nil || c.engine == nil {
		return CacheReport{StrategyHint: "Invalid cache alignment configuration"}
	}

	tokens := c.engine.CountTokens(text)
	blockSize := profile.BlockSize
	if blockSize <= 0 {
		return CacheReport{CurrentTokens: tokens, StrategyHint: "Invalid cache alignment configuration"}
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

	return CacheReport{
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
		StrategyHint:       cacheStrategyHint(tokens, minimum, padding, remainder),
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

func cacheStrategyHint(tokens int, minimum int, padding int, remainder int) string {
	if minimum > 0 && tokens < minimum {
		return fmt.Sprintf("Prompt is %d tokens below the configured cache minimum", minimum-tokens)
	}
	if padding == 0 && remainder == 0 {
		return "Prompt is aligned to the configured cache block boundary"
	}
	return fmt.Sprintf("Prompt is %d tokens from the next configured cache block boundary", padding)
}
