// Package openai owns embedded OpenAI-compatible tokenizer assets.
package openai

import _ "embed"

//go:embed data/cl100k_base.tiktoken
var CL100KBaseData []byte

//go:embed data/o200k_base.tiktoken
var O200KBaseData []byte
