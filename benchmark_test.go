package omnitoken

import (
	"strings"
	"testing"
)

var benchmarkCountSink int
var benchmarkTokenSink []int

var benchmarkTexts = map[string]string{
	"short":   "hello world",
	"json":    "You are a helpful assistant. Summarize this JSON payload, preserve markdown, and explain edge cases: {\"hello\": \"world\", \"n\": 123456}.",
	"unicode": "こんにちは世界 😀 test 中文测试 مرحبا بالعالم snake_case/path-to/file.go",
	"code":    "func main() {\n\tif err := run(context.Background()); err != nil {\n\t\treturn err\n\t}\n\treturn nil\n}",
	"long":    strings.Repeat("System instruction: preserve JSON, markdown, code, Unicode, and exact whitespace. ", 64),
}

func BenchmarkCountTokens(b *testing.B) {
	for _, encoding := range []string{EncodingCL100KBase, EncodingO200KBase, EncodingO200KHarmony} {
		engine, err := ForEncoding(encoding)
		if err != nil {
			b.Fatal(err)
		}
		for name, text := range benchmarkTexts {
			b.Run(encoding+"/"+name, func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(text)))
				for i := 0; i < b.N; i++ {
					benchmarkCountSink = engine.CountTokens(text)
				}
			})
		}
	}
}

func BenchmarkEncodeOrdinary(b *testing.B) {
	for _, encoding := range []string{EncodingCL100KBase, EncodingO200KBase, EncodingO200KHarmony} {
		engine, err := ForEncoding(encoding)
		if err != nil {
			b.Fatal(err)
		}
		for name, text := range benchmarkTexts {
			b.Run(encoding+"/"+name, func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(text)))
				for i := 0; i < b.N; i++ {
					benchmarkTokenSink = engine.EncodeOrdinary(text)
				}
			})
		}
	}
}
