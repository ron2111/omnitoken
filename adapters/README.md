# Adapters

Adapters live outside the root module so OmniToken stays dependency-light for users who only need OpenAI-compatible tokenization.

| Adapter | Package | Scope |
| --- | --- | --- |
| [Gemini](./gemini/README.md) | `github.com/ron2111/omnitoken/adapters/gemini` | Local text tokenization plus API-backed Gemini/Vertex token accounting. |
| [OSS](./oss/README.md) | `github.com/ron2111/omnitoken/adapters/oss` | User-supplied SentencePiece `.model` files with explicit BOS/EOS helpers. |
| [Llama 3](./llama3/README.md) | `github.com/ron2111/omnitoken/adapters/llama3` | User-supplied Llama 3 tiktoken-BPE tokenizer files and known special-token layout. |
| [Mistral](./mistral/README.md) | `github.com/ron2111/omnitoken/adapters/mistral` | User-supplied Mistral Tekken JSON tokenizers; custom patterns guarded by default. |
| [Hugging Face](./huggingface/README.md) | `github.com/ron2111/omnitoken/adapters/huggingface` | WordPiece and simple BPE `tokenizer.json` files. |
| [Anthropic](./anthropic/README.md) | `github.com/ron2111/omnitoken/adapters/anthropic` | API-backed Anthropic Messages token counting and usage parsing. |

Each adapter has its own README with setup, scope, and benchmark notes.

For reference-tokenizer validation, see [adapter parity fixtures](./parity.md).
