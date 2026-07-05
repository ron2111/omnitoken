# OmniToken

Pure-Go LLM tokenizer library for local token counting, encoding, and decoding with OpenAI-compatible BPE (`cl100k_base`, `o200k_base`, `o200k_harmony`) and custom WordPiece and SentencePiece-style vocabularies.

[![Go Reference](https://pkg.go.dev/badge/github.com/ron2111/omnitoken.svg)](https://pkg.go.dev/github.com/ron2111/omnitoken)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)
![Go Version](https://img.shields.io/badge/go-1.24%2B-00ADD8)

OmniToken is useful for prompt sizing, token counting, tokenizer experiments, and cache-boundary planning without CGO, Rust, or Python runtime dependencies.

## Features

- Pure Go tokenizer library for LLM applications.
- OpenAI-compatible BPE token counting for `cl100k_base`, `o200k_base`, and `o200k_harmony`.
- Local `EncodeOrdinary`, `Decode`, and `CountTokens` APIs.
- Custom WordPiece and SentencePiece-style vocabularies.
- Optional adapter modules for Gemini, Llama 3, Mistral, Hugging Face `tokenizer.json`, OSS SentencePiece models, and Anthropic message token counting.

## Install

```powershell
go get github.com/ron2111/omnitoken
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/ron2111/omnitoken"
)

func main() {
	engine, err := omnitoken.ForModel("gpt-4o")
	if err != nil {
		panic(err)
	}

	tokens := engine.EncodeOrdinary("hello world")
	count := engine.CountTokens("hello world")
	text := engine.Decode(tokens)

	fmt.Println(count, text)
}
```

## Support

| Family | Status |
| --- | --- |
| OpenAI `cl100k_base` | Supported |
| OpenAI `o200k_base` | Supported |
| OpenAI `o200k_harmony` | Supported |
| WordPiece local vocabularies | Supported |
| SentencePiece-style local vocabularies | Supported |
| Gemini local text adapter | Optional module |
| OSS SentencePiece adapter | Optional module |
| Llama 3 tiktoken-BPE adapter | Optional module |
| Mistral Tekken adapter | Optional module |
| Hugging Face WordPiece adapter | Optional module |
| Anthropic message counter | Optional module |

## Custom Models

```go
err := omnitoken.RegisterEncoding("my_wordpiece", func() (omnitoken.ModelEngine, error) {
	return omnitoken.NewWordPiece(vocabBytes, omnitoken.WordPieceOptions{Lowercase: true})
})
if err != nil {
	panic(err)
}

err = omnitoken.RegisterModelPrefix("my-model-", "my_wordpiece")
```

## Benchmarks

Recent local sample: Windows amd64, Intel i7-1250U.

| Operation | Encoding | Input | ns/op | B/op | allocs/op |
| --- | --- | --- | ---: | ---: | ---: |
| `CountTokens` | `o200k_base` | JSON | 4,350 | 0 | 0 |
| `EncodeOrdinary` | `o200k_base` | JSON | 6,723 | 448 | 2 |

```powershell
go test ./...
go test -run "^$" -bench "Benchmark" -benchmem -count=1
```

## More

- [Architecture](./docs/architecture.md)
- [Benchmarks and correctness](./docs/benchmarks.md)
- [Adapters](./adapters/README.md)

## Future Scope

- Streaming token counting through the public API.
- More provider adapters with verified local vocab sources.
- Release CI for benchmark regression tracking.

## License

MIT License. See [LICENSE](./LICENSE).
