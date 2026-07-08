package omnitoken

import "errors"

// ErrDisallowedSpecial is returned when Encode sees a known special-token marker
// that was not explicitly allowed.
var ErrDisallowedSpecial = errors.New("omnitoken: disallowed special token")

// EncodeOptions configures special-token handling for Encode.
//
// By default, Encode returns ErrDisallowedSpecial when text contains a known
// special-token marker. Set AllowAllSpecial or list specific markers in
// AllowedSpecial to encode them as special token IDs. Use EncodeOrdinary to
// always treat marker strings as ordinary text.
type EncodeOptions struct {
	AllowAllSpecial bool
	AllowedSpecial  map[string]bool
}

type optionEncoder interface {
	Encode(text string, opts EncodeOptions) ([]int, error)
}

type optionCounter interface {
	CountTokensWithOptions(text string, opts EncodeOptions) (int, error)
}

// Encode encodes text with tiktoken-style special-token handling when the engine
// supports it. Built-in OpenAI BPE engines return ErrDisallowedSpecial for known
// special-token markers unless opts explicitly allows them.
func Encode(engine ModelEngine, text string, opts EncodeOptions) ([]int, error) {
	if engine == nil || text == "" {
		return nil, nil
	}
	if encoder, ok := engine.(optionEncoder); ok {
		return encoder.Encode(text, opts)
	}
	return engine.EncodeOrdinary(text), nil
}

// CountTokensWithOptions counts tokens using the same options as Encode when the
// engine supports special-token handling.
func CountTokensWithOptions(engine ModelEngine, text string, opts EncodeOptions) (int, error) {
	if engine == nil || text == "" {
		return 0, nil
	}
	if counter, ok := engine.(optionCounter); ok {
		return counter.CountTokensWithOptions(text, opts)
	}
	return engine.CountTokens(text), nil
}
