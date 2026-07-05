package mistral

import (
	"os"
	"testing"
)

var countSink int
var tokenSink []int

func BenchmarkMistralTekkenCount(b *testing.B) {
	engine := benchmarkEngine(b)
	for name, text := range benchmarkTexts {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			for i := 0; i < b.N; i++ {
				countSink = engine.CountTokens(text)
			}
		})
	}
}

func BenchmarkMistralTekkenEncode(b *testing.B) {
	engine := benchmarkEngine(b)
	for name, text := range benchmarkTexts {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			for i := 0; i < b.N; i++ {
				tokenSink = engine.EncodeOrdinary(text)
			}
		})
	}
}

func benchmarkEngine(b *testing.B) *Engine {
	b.Helper()
	path := os.Getenv("OMNITOKEN_MISTRAL_TEKKEN_JSON")
	if path == "" {
		b.Skip("set OMNITOKEN_MISTRAL_TEKKEN_JSON")
	}
	engine, err := New(Options{Source: ModelSource{Path: path}})
	if err != nil {
		b.Fatal(err)
	}
	return engine
}

var benchmarkTexts = map[string]string{
	"short":   "hello world",
	"json":    "Summarize this JSON payload and preserve exact fields: {\"hello\":\"world\",\"n\":123456}",
	"unicode": "こんにちは世界 😀 test 中文测试 مرحبا بالعالم",
	"long":    "System instruction: preserve JSON, markdown, code, Unicode, and exact whitespace. System instruction: preserve JSON, markdown, code, Unicode, and exact whitespace.",
}
