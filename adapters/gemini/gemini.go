// Package gemini provides optional local Gemini text-tokenizer registration.
//
// The adapter uses local Gemma SentencePiece model files supplied by the user.
// It is intended for local text-tokenization estimates, not provider billing or
// multimodal accounting parity.
package gemini

import (
	"bytes"
	"errors"
	"sync"

	sentencepiece "github.com/eliben/go-sentencepiece"
	omnitoken "github.com/ron2111/omnitoken"
)

const (
	EncodingGemma2 = "google_gemma2_sentencepiece"
	EncodingGemma3 = "google_gemma3_sentencepiece"
)

// ModelSource points to a local SentencePiece model.
type ModelSource struct {
	Data []byte
	Path string
}

// Options configures local Gemini tokenizer registration.
type Options struct {
	Gemma2 ModelSource
	Gemma3 ModelSource
}

// Engine wraps a local SentencePiece processor as an OmniToken engine.
type Engine struct {
	name string
	proc *sentencepiece.Processor
}

var registration struct {
	sync.Mutex
	registered bool
	options    Options
}

// RegisterWithOptions registers Gemini model mappings backed by local model sources.
func RegisterWithOptions(opts Options) error {
	registration.Lock()
	defer registration.Unlock()
	registration.options = opts
	if registration.registered {
		return nil
	}

	if err := omnitoken.RegisterEncoding(EncodingGemma2, func() (omnitoken.ModelEngine, error) {
		return newEngine(EncodingGemma2, modelSource(EncodingGemma2))
	}); err != nil {
		return err
	}
	if err := omnitoken.RegisterEncoding(EncodingGemma3, func() (omnitoken.ModelEngine, error) {
		return newEngine(EncodingGemma3, modelSource(EncodingGemma3))
	}); err != nil {
		return err
	}

	for _, model := range gemma2Models {
		if err := omnitoken.RegisterProviderModel(omnitoken.ProviderGoogle, model, EncodingGemma2); err != nil {
			return err
		}
	}
	for _, model := range gemma3Models {
		if err := omnitoken.RegisterProviderModel(omnitoken.ProviderGoogle, model, EncodingGemma3); err != nil {
			return err
		}
	}
	registration.registered = true
	return nil
}

func modelSource(encoding string) ModelSource {
	registration.Lock()
	defer registration.Unlock()
	if encoding == EncodingGemma2 {
		return registration.options.Gemma2
	}
	return registration.options.Gemma3
}

func newEngine(name string, source ModelSource) (*Engine, error) {
	proc, err := newProcessor(source)
	if err != nil {
		return nil, err
	}
	return &Engine{name: name, proc: proc}, nil
}

func newProcessor(source ModelSource) (*sentencepiece.Processor, error) {
	if len(source.Data) > 0 {
		return sentencepiece.NewProcessor(bytes.NewReader(source.Data))
	}
	if source.Path != "" {
		return sentencepiece.NewProcessorFromPath(source.Path)
	}
	return nil, errors.New("omnitoken gemini: local SentencePiece model data or path is required")
}

// Encoding returns the adapter encoding name.
func (e *Engine) Encoding() string {
	if e == nil {
		return ""
	}
	return e.name
}

// EncodeOrdinary encodes text with the configured local SentencePiece model.
func (e *Engine) EncodeOrdinary(text string) []int {
	if e == nil || e.proc == nil || text == "" {
		return nil
	}
	tokens := e.proc.Encode(text)
	ids := make([]int, len(tokens))
	for i, token := range tokens {
		ids[i] = token.ID
	}
	return ids
}

// CountTokens returns the number of local SentencePiece tokens in text.
func (e *Engine) CountTokens(text string) int {
	if e == nil || e.proc == nil || text == "" {
		return 0
	}
	return len(e.proc.Encode(text))
}

// Decode decodes token IDs with the configured local SentencePiece model.
func (e *Engine) Decode(tokens []int) string {
	if e == nil || e.proc == nil || len(tokens) == 0 {
		return ""
	}
	return e.proc.Decode(tokens)
}

var gemma2Models = []string{
	"gemini-1.0-pro",
	"gemini-1.5-pro",
	"gemini-1.5-flash",
	"gemini-1.5-flash-8b",
}

var gemma3Models = []string{
	"gemini-2.0-flash",
	"gemini-2.0-flash-lite",
	"gemini-2.5-pro",
	"gemini-2.5-flash",
	"gemini-2.5-flash-lite",
	"gemini-3-pro-preview",
}
