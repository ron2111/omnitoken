package comparegoogle

import (
	"os"
	"testing"

	"github.com/ron2111/omnitoken"
	"github.com/ron2111/omnitoken/adapters/gemini"
	"google.golang.org/genai"
	"google.golang.org/genai/tokenizer"
)

var countSink int

var texts = map[string]string{
	"short":   "hello world",
	"json":    "Summarize this JSON payload and preserve exact fields: {\"hello\":\"world\",\"n\":123456}",
	"unicode": "こんにちは世界 😀 test 中文测试 مرحبا بالعالم",
	"long":    "System instruction: preserve JSON, markdown, code, Unicode, and exact whitespace. System instruction: preserve JSON, markdown, code, Unicode, and exact whitespace.",
}

func TestOmniTokenMatchesGoogleLocalTokenizer(t *testing.T) {
	model := compareModel(t)
	if err := gemini.Register(); err != nil {
		t.Fatal(err)
	}
	engine, err := omnitoken.ForModel(model)
	if err != nil {
		t.Fatal(err)
	}
	tok, err := tokenizer.NewLocalTokenizer(model)
	if err != nil {
		t.Fatal(err)
	}
	for name, text := range texts {
		t.Run(model+"/"+name, func(t *testing.T) {
			content := []*genai.Content{{Parts: []*genai.Part{{Text: text}}}}
			result, err := tok.CountTokens(content, nil)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := engine.CountTokens(text), int(result.TotalTokens); got != want {
				t.Fatalf("CountTokens(%q) = %d, want Google local tokenizer %d", text, got, want)
			}
		})
	}
}

func BenchmarkOmniTokenGemini(b *testing.B) {
	model := compareModel(b)
	if err := gemini.Register(); err != nil {
		b.Fatal(err)
	}
	engine, err := omnitoken.ForModel(model)
	if err != nil {
		b.Fatal(err)
	}
	for name, text := range texts {
		b.Run(model+"/"+name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			for i := 0; i < b.N; i++ {
				countSink = engine.CountTokens(text)
			}
		})
	}
}

func BenchmarkGoogleLocalTokenizer(b *testing.B) {
	model := compareModel(b)
	tok, err := tokenizer.NewLocalTokenizer(model)
	if err != nil {
		b.Fatal(err)
	}
	for name, text := range texts {
		content := []*genai.Content{{Parts: []*genai.Part{{Text: text}}}}
		b.Run(model+"/"+name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			for i := 0; i < b.N; i++ {
				result, err := tok.CountTokens(content, nil)
				if err != nil {
					b.Fatal(err)
				}
				countSink = int(result.TotalTokens)
			}
		})
	}
}

func compareModel(tb testing.TB) string {
	tb.Helper()
	if os.Getenv("OMNITOKEN_GEMINI_COMPARE") == "" {
		tb.Skip("set OMNITOKEN_GEMINI_COMPARE=1")
	}
	model := os.Getenv("OMNITOKEN_GEMINI_COMPARE_MODEL")
	if model == "" {
		model = "gemini-2.5-flash"
	}
	return model
}
