# OmniToken Benchmark Baseline: i7-1250U, 2026-07-05

This baseline was generated with the checkpointed benchmark harness on a Windows 11 machine with an Intel i7-1250U CPU.

## Command

```powershell
.\benchmarks\scripts\run.ps1 -Count 10 -Benchtime 3s -Timeout 90m -DockerGo -Rust
```

## Included Files

| File | Purpose |
| --- | --- |
| `summary.md` | Full human-readable report. |
| `combined.csv` | Normalized benchmark data. |
| `metadata.json` | Machine, Go, Docker, and git metadata. |
| `speedups.svg` | Full speedup chart. |
| `speedups-readme.svg` | Compact README-style speedup chart. |
| `count_ns_per_op.svg` | Count latency chart. |
| `encode_ns_per_op.svg` | Encode latency chart. |
| `decode_ns_per_op.svg` | Decode latency chart. |

## Headline Results

| Comparison | Geomean result |
| --- | ---: |
| OmniToken `CountTokens` vs `tiktoken-go` | 15.84x faster |
| OmniToken `EncodeOrdinary` vs `tiktoken-go` | 13.09x faster |
| OmniToken `Decode` vs `tiktoken-go` | 2.29x faster |
| OmniToken `CountTokens` vs OpenAI Rust `tiktoken` | 0.96x, near parity |
| OmniToken `EncodeOrdinary` vs OpenAI Rust `tiktoken` | 0.75x, competitive pure Go |

Results are machine- and workload-specific. Re-run the benchmark harness on target hardware before making production performance decisions.
