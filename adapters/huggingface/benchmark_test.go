package huggingface

import "testing"

var countSink int
var tokenSink []int

func BenchmarkHuggingFaceWordPieceCount(b *testing.B) {
	engine, err := NewTokenizerJSON([]byte(bertTokenizerJSON), Options{Name: "bench"})
	if err != nil {
		b.Fatal(err)
	}
	texts := map[string]string{
		"short": "Hello world",
		"split": "unaffable unaffable unaffable",
		"mixed": "Hello, world! [MASK] 中文",
	}
	for name, text := range texts {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			for i := 0; i < b.N; i++ {
				countSink = engine.CountTokens(text)
			}
		})
	}
}

func BenchmarkHuggingFaceWordPieceEncode(b *testing.B) {
	engine, err := NewTokenizerJSON([]byte(bertTokenizerJSON), Options{Name: "bench"})
	if err != nil {
		b.Fatal(err)
	}
	texts := map[string]string{
		"short": "Hello world",
		"split": "unaffable unaffable unaffable",
		"mixed": "Hello, world! [MASK] 中文",
	}
	for name, text := range texts {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			for i := 0; i < b.N; i++ {
				tokenSink = engine.EncodeOrdinary(text)
			}
		})
	}
}
