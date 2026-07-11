# Mistral Adapter

Optional local Mistral Tekken tokenization from a user-supplied `tekken.json` file.

```powershell
go get github.com/ron2111/omnitoken/adapters/mistral
```

## Usage

```go
err := mistral.Register(mistral.Options{
	Source: mistral.ModelSource{Path: "./tekken.json"},
})
if err != nil {
	panic(err)
}

err = mistral.RegisterModelPrefixes("mistral-")
if err != nil {
	panic(err)
}

engine, err := omnitoken.ForModel("mistral-local")
```

## Scope

- Supports Mistral Tekken JSON tokenizers supplied by the user.
- Does not bundle Mistral tokenizer files.
- Provides ordinary local text encode/count/decode.
- Uses Tekken `config.pattern` when it is compatible with Go's regexp engine.
- Tekken patterns that use unsupported regexp syntax are rejected by default; set `AllowUnsupportedPattern` only for experimental CL100K-style segmentation fallback.
- Does not claim full Mistral API, tool, multimodal, or billing parity.

## Benchmarks

```powershell
$env:OMNITOKEN_MISTRAL_TEKKEN_JSON="C:\\path\\to\\tekken.json"
go test -run "^$" -bench "BenchmarkMistral" -benchmem -count=1
```
