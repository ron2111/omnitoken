// Package omnitoken provides high-performance tokenization primitives for LLM applications.
package omnitoken

import "errors"

// ErrUnsupportedModel is returned when no tokenizer engine is registered for a model.
var ErrUnsupportedModel = errors.New("omnitoken: unsupported model")

// ErrUnsupportedEncoding is returned when no tokenizer engine is registered for an encoding.
var ErrUnsupportedEncoding = errors.New("omnitoken: unsupported encoding")

// Segmenter abstracts the text-chopping layer away from regular expressions.
type Segmenter interface {
	// Next evaluates a byte slice and returns the ending index of the next text block.
	Next(src []byte, start int) int
}

// ModelEngine defines how token vocabulary and merge sequences are evaluated.
type ModelEngine interface {
	EncodeOrdinary(text string) []int
	Decode(tokens []int) string
	CountTokens(text string) int
}

// StreamCounter counts tokens across incremental input chunks.
type StreamCounter interface {
	Write(chunk []byte) (int, error)
	Count() int
	Reset()
}

// StreamingModelEngine is implemented by engines that can count streamed input.
type StreamingModelEngine interface {
	ModelEngine
	NewStreamCounter() StreamCounter
}
