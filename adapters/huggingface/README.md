# Hugging Face Adapter

Optional Hugging Face `tokenizer.json` loading for OmniToken.

```powershell
go get github.com/ron2111/omnitoken/adapters/huggingface
```

## Scope

- Supports BERT-style WordPiece `tokenizer.json` files.
- Supports simple Hugging Face BPE `tokenizer.json` files with `vocab` plus ordered `merges`; model type can be inferred when `model.type` is omitted.
- Supports ByteLevel BPE pre-tokenization/decoding for GPT-2/RoBERTa-style vocabularies, including `add_prefix_space`.
- Supports ordinary encode/count/decode without adding special post-processing templates.
- Handles `BertNormalizer`, `BertPreTokenizer`, `WordPiece` decoder, simple special added tokens, and WordPiece vocab maps.
- Strict mode rejects unsupported tokenizer components instead of silently claiming parity.
- Does not claim full `AutoTokenizer` parity, offsets, padding, truncation, pair encoding, Unigram tokenizers, post-processing templates, or arbitrary tokenizer pipelines.

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

Use `Options{Permissive: true}` only when you intentionally want to load a tokenizer that contains unsupported metadata fields and accept ordinary-tokenization drift.

## Benchmarks

```powershell
go test -run "^$" -bench "BenchmarkHuggingFace" -benchmem -count=1
```
