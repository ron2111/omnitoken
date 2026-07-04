# Contributing

Thanks for your interest in contributing to OmniToken.

## Development

Run the standard checks before opening a pull request:

```powershell
go test ./...
go vet ./...
go test -run "^$" -bench "Benchmark" -benchmem -count=1
```

## Guidelines

- Keep tokenizer hot paths allocation-conscious.
- Avoid regex in production tokenization paths.
- Add parity tests for tokenizer behavior changes.
- Keep public APIs small and stable.
- Document known incompatibilities or limitations.

## Pull Requests

Please include:

- A short summary of the change.
- Tests or benchmarks for tokenizer behavior changes.
- Any known compatibility risks.
