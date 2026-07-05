# OmniToken

Pure-Go token counting, encoding, and decoding for OpenAI-compatible and custom tokenizer families.

[![Go Reference](https://pkg.go.dev/badge/github.com/ron2111/omnitoken.svg)](https://pkg.go.dev/github.com/ron2111/omnitoken)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)
![Go Version](https://img.shields.io/badge/go-1.24%2B-00ADD8)

OmniToken ships a fast OpenAI BPE tokenizer, a zero-allocation count hot path, prompt cache helpers, and an extensible registry for WordPiece and SentencePiece-style vocabularies.

## Install

```powershell
go get github.com/ron2111/omnitoken
```

## Quick Start

```go
engine, err := omnitoken.ForModel("gpt-4o")
if err != nil {
	panic(err)
}

tokens := engine.EncodeOrdinary("hello world")
count := engine.CountTokens("hello world")
text := engine.Decode(tokens)
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
- [Gemini adapter](./adapters/gemini/README.md)

## Future Scope

- Streaming token counting through the public API.
- Binary SentencePiece `.model` support in the root package.
- More provider adapters with verified local vocab sources.
- Release CI for benchmark regression tracking.

## License

MIT License. See [LICENSE](./LICENSE).
