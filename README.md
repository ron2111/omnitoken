# OmniToken

OmniToken is a pure-Go tokenizer engine for fast token counting, encoding, and decoding of OpenAI-compatible tokenizer families.

The current implementation focuses on OpenAI Phase 1 support: `cl100k_base`, `o200k_base`, and `o200k_harmony`, using embedded `.tiktoken` vocabulary assets, regex-free scanners, and a pure-Go BPE merge loop.

## Features

- Pure Go runtime with no CGO, Rust, Python, or native tokenizer dependency.
- OpenAI model registry through `ForModel`.
- Direct encoding registry through `ForEncoding`.
- Support for `cl100k_base`, `o200k_base`, and `o200k_harmony`.
- Regex-free tokenizer scanners for the hot path.
- Token encode, decode, and count APIs.
- Prompt cache block alignment helper.
- Smoke, edge-case, and 50,000-case parity tests against OpenAI tokenizer outputs.
- Benchmarks with `CountTokens` reaching `0 allocs/op` across the included benchmark matrix.

## Install

```powershell
go get github.com/ron2111/omnitoken
```

## Usage

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

	fmt.Println(tokens)
	fmt.Println(count)
	fmt.Println(text)
}
```

## Supported Encodings

| Encoding | Status | Notes |
| --- | --- | --- |
| `cl100k_base` | Supported | Used by GPT-4, GPT-3.5, and embedding-era models. |
| `o200k_base` | Supported | Used by GPT-4o, GPT-4.1, GPT-5-style, and newer OpenAI models. |
| `o200k_harmony` | Supported | Uses O200K mergeable ranks plus Harmony special-token mappings. |

## Cache Alignment

```go
engine, err := omnitoken.ForEncoding(omnitoken.EncodingO200KBase)
if err != nil {
	panic(err)
}

aligner := omnitoken.NewCacheAligner(engine)
report := aligner.AlignPrompt(systemPrompt, 1024)
```

## Verification

```powershell
go test ./...
go vet ./...
go test -run "^$" -bench "Benchmark" -benchmem -count=1
```

Recent local benchmark sample on Windows amd64, Intel i7-1250U:

```text
BenchmarkCountTokens/o200k_base/json-12          6426 ns/op     0 B/op     0 allocs/op
BenchmarkEncodeOrdinary/o200k_base/json-12       6723 ns/op   448 B/op     2 allocs/op
```

## Current Limitations

- Streaming token counting is not implemented yet.
- Claude, Gemini, SentencePiece, WordPiece, and Llama adapters are not implemented yet.
- CI benchmark regression tracking is not configured yet.
- Race testing requires a local CGO toolchain on Windows.

## License

MIT License. See [LICENSE](./LICENSE).
