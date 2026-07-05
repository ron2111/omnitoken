package gemini

import (
	"os"
	"testing"
)

var benchmarkCountSink int
var benchmarkTokenSink []int
var benchmarkTextSink string

var benchmarkTexts = map[string]string{
	"short":   "hello world",
	"json":    "Summarize this JSON payload and preserve exact fields: {\"hello\":\"world\",\"n\":123456}",
	"unicode": "こんにちは世界 😀 test 中文测试 مرحبا بالعالم",
	"long":    "System instruction: preserve JSON, markdown, code, Unicode, and exact whitespace. System instruction: preserve JSON, markdown, code, Unicode, and exact whitespace.",
}

func BenchmarkGeminiCountTokens(b *testing.B) {
	benchmarkGemini(b, func(b *testing.B, engine *Engine, text string) {
		for i := 0; i < b.N; i++ {
			benchmarkCountSink = engine.CountTokens(text)
		}
	})
}

func BenchmarkGeminiEncodeOrdinary(b *testing.B) {
	benchmarkGemini(b, func(b *testing.B, engine *Engine, text string) {
		for i := 0; i < b.N; i++ {
			benchmarkTokenSink = engine.EncodeOrdinary(text)
		}
	})
}

func BenchmarkGeminiDecode(b *testing.B) {
	benchmarkGemini(b, func(b *testing.B, engine *Engine, text string) {
		tokens := engine.EncodeOrdinary(text)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchmarkTextSink = engine.Decode(tokens)
		}
	})
}

func benchmarkGemini(b *testing.B, run func(*testing.B, *Engine, string)) {
	for _, tc := range []struct {
		name     string
		encoding string
		env      string
	}{
		{"gemma2", EncodingGemma2, "OMNITOKEN_GEMINI_GEMMA2_MODEL"},
		{"gemma3", EncodingGemma3, "OMNITOKEN_GEMINI_GEMMA3_MODEL"},
	} {
		path := os.Getenv(tc.env)
		if path == "" {
			continue
		}
		engine, err := newEngine(tc.encoding, Options{}, ModelSource{Path: path})
		if err != nil {
			b.Fatal(err)
		}
		for name, text := range benchmarkTexts {
			b.Run(tc.name+"/"+name, func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(text)))
				run(b, engine, text)
			})
		}
	}
	if os.Getenv("OMNITOKEN_GEMINI_GEMMA2_MODEL") == "" && os.Getenv("OMNITOKEN_GEMINI_GEMMA3_MODEL") == "" {
		b.Skip("set OMNITOKEN_GEMINI_GEMMA2_MODEL or OMNITOKEN_GEMINI_GEMMA3_MODEL")
	}
}
