# Hugging Face Adapter

Optional Hugging Face `tokenizer.json` loading for OmniToken.

```powershell
go get github.com/ron2111/omnitoken/adapters/huggingface
```

## Scope

- Supports BERT-style WordPiece `tokenizer.json` files.
- Supports ordinary encode/count/decode without adding special post-processing templates.
- Handles `BertNormalizer`, `BertPreTokenizer`, `WordPiece` decoder, simple special added tokens, and WordPiece vocab maps.
- Does not claim full `AutoTokenizer` parity, offsets, padding, truncation, pair encoding, ByteLevel BPE, or arbitrary tokenizer pipelines.

## Usage

```go
data, err := os.ReadFile("tokenizer.json")
if err != nil {
	panic(err)
}

err = huggingface.RegisterTokenizerJSON("my_bert", data, huggingface.Options{})
if err != nil {
	panic(err)
}

err = omnitoken.RegisterModel("my-bert-model", "my_bert")
if err != nil {
	panic(err)
}

engine, err := omnitoken.ForModel("my-bert-model")
```

## Benchmarks

```powershell
go test -run "^$" -bench "BenchmarkHuggingFace" -benchmem -count=1
```
