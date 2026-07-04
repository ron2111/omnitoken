package omnitoken

import _ "embed"

//go:embed internal/openai/data/cl100k_base.tiktoken
var cl100kBaseData []byte

//go:embed internal/openai/data/o200k_base.tiktoken
var o200kBaseData []byte
