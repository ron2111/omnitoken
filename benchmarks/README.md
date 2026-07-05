# Benchmarks

OmniToken benchmarks compare count, encode, and decode behavior across comparable tokenizer implementations.

## Runners

| Runner | Scope |
| --- | --- |
| `omnitoken` | Root Go implementation. |
| `tiktoken-go` | Go OpenAI tokenizer baseline. |
| `openai-rust` | Dockerized OpenAI `tiktoken` Rust core. |

## Run

```powershell
.\benchmarks\scripts\run.ps1 -Count 3 -Benchtime 1s
```

Run Rust comparison too:

```powershell
.\benchmarks\scripts\run.ps1 -Count 3 -Benchtime 1s -Rust
```

## Output

```text
benchmarks/results/go.raw.txt
benchmarks/results/go.jsonl
benchmarks/results/rust.csv
benchmarks/results/combined.csv
benchmarks/reports/summary.md
benchmarks/reports/*.svg
```

## Caveats

- `CountTokens` is a native OmniToken count-only operation.
- Some competitors count via `len(Encode(...))`; those rows are labeled `count_by_encode`.
- Dockerized Rust results are reproducible but not identical to native Windows Go timing.
- Tokenizer parity must pass before treating speed comparisons as meaningful.
