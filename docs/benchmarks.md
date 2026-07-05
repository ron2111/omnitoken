# Benchmarks and Correctness

## Correctness

OpenAI-compatible outputs are checked with layered tests.

| Check | Purpose |
| --- | --- |
| Smoke fixtures | Known token IDs for common strings, Unicode, code, JSON, and whitespace. |
| Edge cases | Emoji, CJK, punctuation, repeated spaces, and special-token text. |
| Reference corpus | 50,000 deterministic cases checked against expected OpenAI tokenizer outputs. |

For supported OpenAI encodings, correctness means identical token ID sequences for the same input text.

## Measured Results

Measured after scanner/decode optimizations on Windows 11 amd64, Intel i7-1250U, Go 1.24.2.

| Operation | Encoding | Input | Typical ns/op | B/op | allocs/op |
| --- | --- | --- | ---: | ---: | ---: |
| `CountTokens` | `cl100k_base` | JSON | 1,517 | 0 | 0 |
| `EncodeOrdinary` | `cl100k_base` | JSON | 1,661 | 288 | 1 |
| `Decode` | `cl100k_base` | JSON | 204 | 288 | 2 |
| `CountTokens` | `o200k_base` | JSON | 2,152 | 0 | 0 |
| `EncodeOrdinary` | `o200k_base` | JSON | 1,835 | 288 | 1 |
| `Decode` | `o200k_base` | JSON | 192 | 288 | 2 |

Latest completed comparison report:

| Comparison | Geomean speedup |
| --- | ---: |
| OmniToken `CountTokens` vs `tiktoken-go` count-by-encode | 15.84x |
| OmniToken `EncodeOrdinary` vs `tiktoken-go` encode | 13.09x |
| OmniToken `Decode` vs `tiktoken-go` decode | 2.29x |
| OmniToken `CountTokens` vs OpenAI Rust `tiktoken` count-by-encode | 0.96x |
| OmniToken `EncodeOrdinary` vs OpenAI Rust `tiktoken` encode | 0.75x |

These numbers are workload- and machine-specific. The checkpointed benchmark harness records machine metadata and should be rerun on target hardware for formal claims.

Curated baseline report: [`benchmarks/baselines/i7-1250u-2026-07-05/summary.md`](../benchmarks/baselines/i7-1250u-2026-07-05/summary.md).

## Benchmark Commands

For the full comparison harness with `tiktoken-go` and Dockerized OpenAI Rust `tiktoken`, see [`benchmarks/README.md`](../benchmarks/README.md).

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
