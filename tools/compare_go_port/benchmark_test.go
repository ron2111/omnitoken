package comparegoport

import (
	"strings"
	"testing"

	tiktoken "github.com/pkoukk/tiktoken-go"
	omnitoken "github.com/ron2111/omnitoken"
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

func BenchmarkOmiTokenCountTokens(b *testing.B) {
	for _, encoding := range []string{omnitoken.EncodingCL100KBase, omnitoken.EncodingO200KBase} {
		engine, err := omnitoken.ForEncoding(encoding)
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

func BenchmarkGoPortCountByEncode(b *testing.B) {
	for _, encoding := range []string{omnitoken.EncodingCL100KBase, omnitoken.EncodingO200KBase} {
		engine, err := tiktoken.GetEncoding(encoding)
		if err != nil {
			b.Fatal(err)
		}
		for name, text := range benchmarkTexts {
			b.Run(encoding+"/"+name, func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(text)))
				for i := 0; i < b.N; i++ {
					benchmarkCountSink = len(engine.EncodeOrdinary(text))
				}
			})
		}
	}
}

func BenchmarkOmiTokenEncodeOrdinary(b *testing.B) {
	for _, encoding := range []string{omnitoken.EncodingCL100KBase, omnitoken.EncodingO200KBase} {
		engine, err := omnitoken.ForEncoding(encoding)
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

func BenchmarkGoPortEncodeOrdinary(b *testing.B) {
	for _, encoding := range []string{omnitoken.EncodingCL100KBase, omnitoken.EncodingO200KBase} {
		engine, err := tiktoken.GetEncoding(encoding)
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
