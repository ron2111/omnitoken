package omnitoken

import (
	"fmt"
	"strings"
	"sync"
)

const (
	EncodingCL100KBase   = "cl100k_base"
	EncodingO200KBase    = "o200k_base"
	EncodingO200KHarmony = "o200k_harmony"
)

var exactModelEncodings = map[string]string{
	"o1":                     EncodingO200KBase,
	"o3":                     EncodingO200KBase,
	"o4-mini":                EncodingO200KBase,
	"gpt-5":                  EncodingO200KBase,
	"gpt-4.1":                EncodingO200KBase,
	"gpt-4o":                 EncodingO200KBase,
	"gpt-4":                  EncodingCL100KBase,
	"gpt-3.5":                EncodingCL100KBase,
	"gpt-3.5-turbo":          EncodingCL100KBase,
	"gpt-35-turbo":           EncodingCL100KBase,
	"davinci-002":            EncodingCL100KBase,
	"babbage-002":            EncodingCL100KBase,
	"text-embedding-ada-002": EncodingCL100KBase,
	"text-embedding-3-small": EncodingCL100KBase,
	"text-embedding-3-large": EncodingCL100KBase,
}

var prefixModelEncodings = []struct {
	prefix   string
	encoding string
}{
	{"o1-", EncodingO200KBase},
	{"o3-", EncodingO200KBase},
	{"o4-mini-", EncodingO200KBase},
	{"gpt-5-", EncodingO200KBase},
	{"gpt-4.5-", EncodingO200KBase},
	{"gpt-4.1-", EncodingO200KBase},
	{"chatgpt-4o-", EncodingO200KBase},
	{"gpt-4o-", EncodingO200KBase},
	{"gpt-4-", EncodingCL100KBase},
	{"gpt-3.5-turbo-", EncodingCL100KBase},
	{"gpt-35-turbo-", EncodingCL100KBase},
	{"gpt-oss-", EncodingO200KHarmony},
	{"ft:gpt-4o", EncodingO200KBase},
	{"ft:gpt-4", EncodingCL100KBase},
	{"ft:gpt-3.5-turbo", EncodingCL100KBase},
	{"ft:davinci-002", EncodingCL100KBase},
	{"ft:babbage-002", EncodingCL100KBase},
}

var engineCache sync.Map

// ForModel resolves a model name to a tokenizer engine.
func ForModel(model string) (ModelEngine, error) {
	encoding, ok := encodingForModel(model)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedModel, model)
	}
	return ForEncoding(encoding)
}

// ForEncoding resolves an OpenAI encoding name to a tokenizer engine.
func ForEncoding(encoding string) (ModelEngine, error) {
	if cached, ok := engineCache.Load(encoding); ok {
		return cached.(*Engine), nil
	}

	engine, err := buildEncoding(encoding)
	if err != nil {
		return nil, err
	}
	actual, _ := engineCache.LoadOrStore(encoding, engine)
	return actual.(*Engine), nil
}

func encodingForModel(model string) (string, bool) {
	if encoding, ok := exactModelEncodings[model]; ok {
		return encoding, true
	}
	for _, entry := range prefixModelEncodings {
		if strings.HasPrefix(model, entry.prefix) {
			return entry.encoding, true
		}
	}
	return "", false
}
