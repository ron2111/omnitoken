package oss

import (
	"os"
	"testing"
)

var countSink int
var tokenSink []int
var textSink string

var benchTexts = map[string]string{
	"short":   "hello world",
	"json":    "Summarize this JSON payload and preserve exact fields: {\"hello\":\"world\",\"n\":123456}",
	"unicode": "こんにちは世界 😀 test 中文测试 مرحبا بالعالم",
	"long":    "System instruction: preserve JSON, markdown, code, Unicode, and exact whitespace. System instruction: preserve JSON, markdown, code, Unicode, and exact whitespace.",
}

func BenchmarkOSSCountTokens(b *testing.B) {
	benchmarkOSS(b, func(b *testing.B, engine *Engine, text string) {
		for i := 0; i < b.N; i++ {
			countSink = engine.CountTokens(text)
		}
	})
}

func BenchmarkOSSEncodeOrdinary(b *testing.B) {
	benchmarkOSS(b, func(b *testing.B, engine *Engine, text string) {
		for i := 0; i < b.N; i++ {
			tokenSink = engine.EncodeOrdinary(text)
		}
	})
}

func BenchmarkOSSDecode(b *testing.B) {
	benchmarkOSS(b, func(b *testing.B, engine *Engine, text string) {
		tokens := engine.EncodeOrdinary(text)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			textSink = engine.Decode(tokens)
		}
	})
}

func benchmarkOSS(b *testing.B, run func(*testing.B, *Engine, string)) {
	path := os.Getenv("OMNITOKEN_OSS_SENTENCEPIECE_MODEL")
	if path == "" {
		b.Skip("set OMNITOKEN_OSS_SENTENCEPIECE_MODEL")
	}
	engine, err := NewSentencePiece("oss_benchmark", Options{Source: ModelSource{Path: path}})
	if err != nil {
		b.Fatal(err)
	}
	for name, text := range benchTexts {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			run(b, engine, text)
		})
	}
}
