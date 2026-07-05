# Benchmarks and Correctness

## Correctness

OpenAI-compatible outputs are checked with layered tests.

| Check | Purpose |
| --- | --- |
| Smoke fixtures | Known token IDs for common strings, Unicode, code, JSON, and whitespace. |
| Edge cases | Emoji, CJK, punctuation, repeated spaces, and special-token text. |
| Reference corpus | 50,000 deterministic cases checked against expected OpenAI tokenizer outputs. |

For supported OpenAI encodings, correctness means identical token ID sequences for the same input text.

## Benchmark Commands

```powershell
go test ./...
go vet ./...
go test -run "^$" -bench "Benchmark" -benchmem -count=1
```

Adapter modules are tested from their own directories:

```powershell
Push-Location adapters/huggingface
go test ./...
go test -run "^$" -bench "BenchmarkHuggingFace" -benchmem -count=1
Pop-Location
```

Comparison against `tiktoken-go` lives in a separate tool module so normal users do not inherit benchmark dependencies.

```powershell
Push-Location tools/compare_go_port
go test -run "^$" -bench "Benchmark" -benchmem -count=1
Pop-Location
```

OpenAI's Rust-backed Python package can be checked with:

```powershell
python -m pip install tiktoken
python tools/openai_reference_benchmark.py
```

## Reading Results

| Metric | Meaning |
| --- | --- |
| `ns/op` | Nanoseconds per operation. Lower is faster. |
| `B/op` | Heap bytes allocated per operation. |
| `allocs/op` | Heap allocation events per operation. |
| `MB/s` | Input throughput from `b.SetBytes`. |

`CountTokens` is designed to count without materializing the final token slice, so supported OpenAI count benchmarks target `0 B/op` and `0 allocs/op`.
