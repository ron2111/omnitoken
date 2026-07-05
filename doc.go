// Package omnitoken provides a pure-Go LLM tokenizer library for local token
// counting, encoding, and decoding.
//
// The root module includes OpenAI-compatible byte-pair encoding for cl100k_base,
// o200k_base, and o200k_harmony, plus lightweight WordPiece and
// SentencePiece-style engines for custom vocabularies. Optional adapter modules
// support Gemini, Llama 3, Mistral, Hugging Face tokenizer.json files, OSS
// SentencePiece models, and Anthropic server-side message token counting.
package omnitoken
