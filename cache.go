package omnitoken

import "fmt"

// CacheReport describes how close a prompt is to a provider cache block boundary.
type CacheReport struct {
	CurrentTokens int
	NextBlockSize int
	PaddingNeeded int
	StrategyHint  string
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
	if c == nil || c.engine == nil || providerBlockSize <= 0 {
		return CacheReport{StrategyHint: "Invalid cache alignment configuration"}
	}

	tokens := c.engine.CountTokens(text)
	remainder := tokens % providerBlockSize
	if remainder == 0 {
		return CacheReport{tokens, tokens, 0, "Perfect alignment"}
	}

	needed := providerBlockSize - remainder
	return CacheReport{
		CurrentTokens: tokens,
		NextBlockSize: tokens + needed,
		PaddingNeeded: needed,
		StrategyHint:  fmt.Sprintf("Inject %d tokens of safe system metadata or whitespace padding", needed),
	}
}
