# Gemini Adapter

Optional local Gemini text tokenization for OmniToken.

```powershell
go get github.com/ron2111/omnitoken/adapters/gemini
```

The adapter mirrors the Gemini-to-Gemma local tokenizer mapping used by Google's Go GenAI local tokenizer and wraps Gemma SentencePiece model files as OmniToken engines.

## Usage

```go
import (
	"github.com/ron2111/omnitoken"
	"github.com/ron2111/omnitoken/adapters/gemini"
)

func main() {
	if err := gemini.Register(); err != nil {
		panic(err)
	}

	engine, err := omnitoken.ForModel("gemini-2.5-flash")
	if err != nil {
		panic(err)
	}

	count := engine.CountTokens("hello world")
	_ = count
}
```

`gemini.Register()` uses official Gemma model URLs and SHA-256 hashes, then caches the model file on first use.

For offline or pinned builds, provide local files:

```go
err := gemini.RegisterWithOptions(gemini.Options{
	Offline: true,
	Gemma2: gemini.ModelSource{Path: "./tokenizer.model"},
	Gemma3: gemini.ModelSource{Path: "./gemma3_cleaned_262144_v2.spiece.model"},
})
```

## Scope

- Local text tokenization estimates for supported Gemini model names.
- Same Gemma2/Gemma3 SentencePiece artifacts and exact model mappings as Google's Go local tokenizer source at the time of implementation.
- Not billing-grade multimodal accounting.
- Not a replacement for provider `countTokens`, `computeTokens`, or response usage metadata.

## Benchmarks

Run adapter benchmarks with local model files:

```powershell
$env:OMNITOKEN_GEMINI_GEMMA2_MODEL="C:\\path\\to\\tokenizer.model"
$env:OMNITOKEN_GEMINI_GEMMA3_MODEL="C:\\path\\to\\gemma3_cleaned_262144_v2.spiece.model"
go test -run "^$" -bench "BenchmarkGemini" -benchmem -count=1
```

Run comparison benchmarks against Google's local tokenizer from the comparison module:

```powershell
Push-Location compare_google
$env:OMNITOKEN_GEMINI_COMPARE="1"
go test -run "^$" -bench "Benchmark" -benchmem -count=1
Pop-Location
```
