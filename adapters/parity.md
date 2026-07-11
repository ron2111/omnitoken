# Adapter Parity Fixtures

Adapter parity tests are skipped by default because reference tokenizer assets and generated fixtures are usually large, licensed separately, or require provider tooling.

Each local tokenizer adapter supports newline-delimited JSON fixtures with this shape:

```json
{"name":"ascii","text":"hello world","tokens":[1,2],"decode":"hello world"}
{"name":"count-only","text":"hello world","count":2}
```

Fields:

- `name`: optional test name.
- `text`: input text.
- `tokens`: optional expected ordinary token IDs.
- `count`: optional expected token count. If omitted and `tokens` exists, the count expectation is `len(tokens)`.
- `decode`: optional expected decoded text for `tokens`.

## Environment Variables

### Gemini

```powershell
$env:OMNITOKEN_GEMINI_GEMMA2_MODEL="C:\path\to\tokenizer.model"
$env:OMNITOKEN_GEMINI_GEMMA2_PARITY_JSONL="C:\path\to\gemma2-parity.jsonl"

$env:OMNITOKEN_GEMINI_GEMMA3_MODEL="C:\path\to\gemma3_cleaned_262144_v2.spiece.model"
$env:OMNITOKEN_GEMINI_GEMMA3_PARITY_JSONL="C:\path\to\gemma3-parity.jsonl"
go test ./adapters/gemini
```

For direct comparison against Google's local tokenizer, use the separate comparison module:

```powershell
Push-Location adapters/gemini/compare_google
$env:OMNITOKEN_GEMINI_COMPARE="1"
$env:OMNITOKEN_GEMINI_COMPARE_MODEL="gemini-2.5-flash"
go test ./...
Pop-Location
```

### Hugging Face

```powershell
$env:OMNITOKEN_HUGGINGFACE_TOKENIZER_JSON="C:\path\to\tokenizer.json"
$env:OMNITOKEN_HUGGINGFACE_PARITY_JSONL="C:\path\to\hf-parity.jsonl"
go test ./adapters/huggingface
```

Set `OMNITOKEN_HUGGINGFACE_PERMISSIVE=1` only when intentionally testing ordinary-tokenization drift for tokenizer files with unsupported metadata.

### Llama 3

```powershell
$env:OMNITOKEN_LLAMA3_TOKENIZER_MODEL="C:\path\to\tokenizer.model"
$env:OMNITOKEN_LLAMA3_VARIANT="llama3.1"
$env:OMNITOKEN_LLAMA3_PARITY_JSONL="C:\path\to\llama3-parity.jsonl"
go test ./adapters/llama3
```

`OMNITOKEN_LLAMA3_VARIANT` accepts `llama3`, `llama3.1`, or `llama3.2`.

### Mistral

```powershell
$env:OMNITOKEN_MISTRAL_TEKKEN_JSON="C:\path\to\tekken.json"
$env:OMNITOKEN_MISTRAL_PARITY_JSONL="C:\path\to\mistral-parity.jsonl"
go test ./adapters/mistral
```

Tekken files with `config.pattern` use that pattern when it is compatible with Go's regexp engine. Set `OMNITOKEN_MISTRAL_ALLOW_UNSUPPORTED_PATTERN=1` only to test the experimental CL100K-style fallback for unsupported patterns.

### OSS SentencePiece

```powershell
$env:OMNITOKEN_OSS_SENTENCEPIECE_MODEL="C:\path\to\tokenizer.model"
$env:OMNITOKEN_OSS_PARITY_JSONL="C:\path\to\oss-parity.jsonl"
go test ./adapters/oss
```

## Current Status

| Adapter | Local parity status |
| --- | --- |
| Gemini | Local text tokenizer implemented; fixture parity and Google local-tokenizer comparison are env-gated. Multimodal/accounting uses provider APIs. |
| Hugging Face | WordPiece, simple BPE, and ByteLevel BPE are supported. Full `AutoTokenizer`, Unigram, offsets, post-processors, truncation, padding, and pair encoding are not complete. |
| Llama 3 | Local tiktoken-BPE path implemented with known special-token layout. Needs official fixture runs for each variant. |
| Mistral | Tekken JSON loading implemented. `config.pattern` is used when Go-compatible; unsupported regex syntax is guarded unless explicitly allowed as an experimental fallback. |
| OSS | Generic SentencePiece wrapper with fixture parity support. It delegates to `go-sentencepiece`; OmniToken does not own the full optimized SentencePiece runtime here. |
| Anthropic | No local Claude tokenizer parity. Current Claude tokenizer specs/artifacts are not public; use provider-side counting and final usage parsing. |
