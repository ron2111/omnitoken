# Adapters

Adapters live outside the root module so OmniToken stays dependency-light for users who only need OpenAI-compatible tokenization.

| Adapter | Package | Scope |
| --- | --- | --- |
| Gemini | `github.com/ron2111/omnitoken/adapters/gemini` | Local text estimates using Gemma SentencePiece models. |
| OSS | `github.com/ron2111/omnitoken/adapters/oss` | User-supplied SentencePiece `.model` files for Llama 2-class, Gemma, and similar tokenizers. |
| Hugging Face | `github.com/ron2111/omnitoken/adapters/huggingface` | BERT-style WordPiece `tokenizer.json` files. |
| Anthropic | `github.com/ron2111/omnitoken/adapters/anthropic` | API-backed Anthropic Messages token counting. |

Each adapter has its own README with setup, scope, and benchmark notes.
