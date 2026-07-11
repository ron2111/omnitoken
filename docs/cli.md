# Omni CLI

OmniToken ships a small general-purpose CLI under `cmd/omni`.

## Install

```powershell
go install github.com/ron2111/omnitoken/cmd/omni@latest
```

## Count

```powershell
omni count -model gpt-4o "hello world"
omni count -encoding o200k_base "hello world"
```

## Encode

```powershell
omni encode -encoding o200k_base "hello world"
```

Output:

```json
[24912,2375]
```

## Decode

```powershell
omni decode -encoding o200k_base "24912 2375"
```

## Cache Planning

```powershell
omni cache -model gpt-4o -profile openai "hello world"
```

Analyze a JSONL trace for stable token prefixes and cache-breaker hints:

```powershell
omni cache-sim -model gpt-4o -profile openai -input prompts.jsonl -breakers
```

See [`cacheflow`](../cacheflow/README.md) for the JSONL format.

## Registry Discovery

List encodings registered in the current CLI build:

```powershell
omni encodings
omni encodings -json
```

List exact model mappings:

```powershell
omni models
omni models -json
```

Include prefix mappings such as `gpt-4o-` and fine-tuned model prefixes:

```powershell
omni models -prefixes
```

## Benchmark Integration

The `bench` subcommand runs timing loops inside the Go process and writes Systemcluster-style timing files.

```powershell
omni bench -name "omnitoken - cl100k - small" -model cl100k_base -input data.txt -timings timings -iters 100 -warmup 10 -batch 10
```
