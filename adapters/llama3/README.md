# Llama 3 Adapter

Optional local Llama 3 tiktoken-BPE tokenization from a user-supplied Meta tokenizer file.

```powershell
go get github.com/ron2111/omnitoken/adapters/llama3
```

## Usage

```go
err := llama3.Register(llama3.Options{
	Source: llama3.ModelSource{Path: "./tokenizer.model"},
	Variant: llama3.VariantLlama31,
})
if err != nil {
	panic(err)
}

err = llama3.RegisterModelPrefixes("llama-3-")
if err != nil {
	panic(err)
}

engine, err := omnitoken.ForModel("llama-3-local")
```

## Scope

- Supports Llama 3-family tiktoken-BPE tokenizer files supplied by the user.
- Does not bundle Meta tokenizer files.
- Provides ordinary encode/count/decode plus explicit BOS/EOS helpers.
- Does not claim hosted API billing or chat-template parity.

## Benchmarks

```powershell
$env:OMNITOKEN_LLAMA3_TOKENIZER_MODEL="C:\\path\\to\\tokenizer.model"
go test -run "^$" -bench "BenchmarkLlama3" -benchmem -count=1
```
