# Adapters

Adapters live outside the root module so OmniToken stays dependency-light for users who only need OpenAI-compatible tokenization.

| Adapter | Package | Scope |
| --- | --- | --- |
| [Gemini](./gemini/README.md) | `github.com/ron2111/omnitoken/adapters/gemini` | Local text estimates using Gemma SentencePiece models. |
| [OSS](./oss/README.md) | `github.com/ron2111/omnitoken/adapters/oss` | User-supplied SentencePiece `.model` files for Llama 2-class, Gemma, and similar tokenizers. |
| [Llama 3](./llama3/README.md) | `github.com/ron2111/omnitoken/adapters/llama3` | User-supplied Llama 3 tiktoken-BPE tokenizer files. |
| [Mistral](./mistral/README.md) | `github.com/ron2111/omnitoken/adapters/mistral` | User-supplied Mistral Tekken JSON tokenizers. |
| [Hugging Face](./huggingface/README.md) | `github.com/ron2111/omnitoken/adapters/huggingface` | BERT-style WordPiece `tokenizer.json` files. |
| [Anthropic](./anthropic/README.md) | `github.com/ron2111/omnitoken/adapters/anthropic` | API-backed Anthropic Messages token counting. |

Each adapter has its own README with setup, scope, and benchmark notes.
