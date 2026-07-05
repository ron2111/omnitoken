# Cache Alignment

OmniToken includes a small cache-planning helper for prompt-cache block boundaries.

It does not edit prompts. It counts tokens, calculates boundary distance, and returns a report that users can apply in their own prompt-building code.

## Usage

```go
engine, err := omnitoken.ForModel("gpt-4o")
if err != nil {
	panic(err)
}

aligner := omnitoken.NewCacheAligner(engine)
report := aligner.AlignPromptToProfile(systemPrompt, omnitoken.CacheProfileOpenAI)
```

## Report

| Field | Meaning |
| --- | --- |
| `CurrentTokens` | Token count for the input text. |
| `BlockSize` | Cache block size used for the calculation. |
| `MinimumTokens` | Minimum token threshold in the selected profile. |
| `PreviousBlockSize` | Previous block boundary. |
| `NextBlockSize` | Next block boundary. |
| `Remainder` | Tokens past the previous boundary. |
| `PaddingNeeded` | Additional tokens needed to reach the next boundary. |
| `TokensUntilMinimum` | Tokens needed before the prompt meets the profile minimum. |
| `IsAligned` | Whether the prompt is already on a boundary. |
| `IsEligible` | Whether the prompt meets the profile minimum. |
| `StrategyHint` | Human-readable planning hint. |

## Profiles

```go
omnitoken.CacheProfileGeneric
omnitoken.CacheProfileOpenAI
```

Profiles are local planning helpers, not billing guarantees. Provider cache behavior can change, and final usage should still be checked against provider usage metadata.

## Custom Blocks

For a custom block size:

```go
report := aligner.AlignPrompt(prompt, 1024)
```

For a custom profile:

```go
report := aligner.AlignPromptToProfile(prompt, omnitoken.CacheProfile{
	Name:          "custom",
	BlockSize:     512,
	MinimumTokens: 2048,
})
```
