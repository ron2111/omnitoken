# OSS Adapter

Optional local SentencePiece tokenization for user-supplied OSS model files such as Llama 2-class, Gemma, and older Mistral-style `.model` tokenizers.

```powershell
go get github.com/ron2111/omnitoken/adapters/oss
```

## Usage

```go
err := oss.RegisterSentencePiece("my_llama2", oss.Options{
	Source: oss.ModelSource{Path: "./tokenizer.model"},
})
if err != nil {
	panic(err)
}

err = oss.RegisterModelPrefixes(omnitoken.ProviderMeta, "my_llama2", "llama-2-")
if err != nil {
	panic(err)
}

engine, err := omnitoken.ForModel("llama-2-local")
```

## Scope

- Supports local SentencePiece `.model` files supplied by the user.
- Provides ordinary encode/count/decode through OmniToken's `ModelEngine` API.
- Supports explicit BOS/EOS accounting through `Encode` and `Count`, including empty prompts when the model defines BOS/EOS IDs.
- Does not claim Llama 3 tiktoken-BPE, Mistral Tekken, chat-template, or billing parity.

## Benchmarks

```powershell
$env:OMNITOKEN_OSS_SENTENCEPIECE_MODEL="C:\\path\\to\\tokenizer.model"
go test -run "^$" -bench "BenchmarkOSS" -benchmem -count=1
```
