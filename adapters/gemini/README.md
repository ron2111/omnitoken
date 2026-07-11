# Gemini Adapter

Optional Gemini tokenization and accounting for OmniToken.

```powershell
go get github.com/ron2111/omnitoken/adapters/gemini
```

The adapter has two layers:

- Local text tokenization using Gemma/Gemini SentencePiece model files.
- API-backed Gemini/Vertex `countTokens` accounting for structured, multimodal request payloads.

The local tokenizer mirrors the Gemini-to-Gemma local tokenizer mapping used by Google's Go GenAI local tokenizer and wraps Gemma SentencePiece model files as OmniToken engines.

## Local Text Tokenization

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

## Provider Token Accounting

Use `Client` when you need request-shaped accounting for Gemini content, system instructions, tools, cached content, or multimodal parts.

Gemini Developer API with an API key:

```go
client := gemini.Client{APIKey: apiKey}

result, err := client.CountContentTokens(ctx, "gemini-2.5-flash", []gemini.Content{{
	Role: "user",
	Parts: []gemini.Part{
		{Text: "Describe this image."},
		{InlineData: gemini.InlineDataFromBytes("image/png", imageBytes)},
	},
}})
if err != nil {
	panic(err)
}

fmt.Println(result.TotalTokens)
```

Full generation-shaped request:

```go
result, err := client.CountGenerateContentRequest(ctx, "gemini-2.5-flash", gemini.GenerateContentRequest{
	SystemInstruction: &gemini.Content{Parts: []gemini.Part{{Text: "Answer briefly."}}},
	Contents: []gemini.Content{{
		Role:  "user",
		Parts: []gemini.Part{{Text: "Summarize the attached document."}},
	}},
	Tools: []map[string]any{
		{"functionDeclarations": []any{}},
	},
})
```

Vertex AI publisher model endpoint:

```go
client := gemini.Client{
	BearerToken: accessToken,
	Project:     "my-project",
	Location:    "us-central1",
}

result, err := client.CountContentTokens(ctx, "gemini-2.5-flash", contents)
```

After a real generation call, parse final usage metadata from the provider response:

```go
usage, err := gemini.ParseUsageMetadata(responseJSON)
if err != nil {
	panic(err)
}

fmt.Println(usage.PromptTokenCount, usage.CandidatesTokenCount, usage.TotalTokenCount)
```

## Scope

- Local text tokenization estimates for supported Gemini model names.
- API-backed `countTokens` preflight accounting for Gemini Developer API and Vertex AI publisher model endpoints.
- Structured request support for contents, system instructions, tools, tool config, safety settings, generation config, cached content, inline data, file data, and common multimodal parts.
- Final usage parsing through `UsageMetadata` from generation responses.
- Same Gemma2/Gemma3 SentencePiece artifacts and exact model mappings as Google's Go local tokenizer source at the time of implementation.
- Provider `countTokens` is still preflight input accounting. For multimodal inputs, Google documents it as an estimate that can differ from final consumed tokens.
- Final generation response `usageMetadata` remains authoritative for post-execution usage and billing reconciliation.
- The local tokenizer is not billing-grade multimodal accounting and does not replace provider-side accounting.

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
