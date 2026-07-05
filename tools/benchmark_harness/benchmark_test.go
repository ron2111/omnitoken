package benchmarkharness

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	tiktoken "github.com/pkoukk/tiktoken-go"
	omnitoken "github.com/ron2111/omnitoken"
)

type corpusCase struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Repeat int    `json:"repeat"`
}

var countSink int
var tokenSink []int
var textSink string

func BenchmarkTokenizer(b *testing.B) {
	cases := loadCorpus(b)
	for _, encoding := range []string{omnitoken.EncodingCL100KBase, omnitoken.EncodingO200KBase} {
		omni, err := omnitoken.ForEncoding(encoding)
		if err != nil {
			b.Fatal(err)
		}
		goPort, err := tiktoken.GetEncoding(encoding)
		if err != nil {
			b.Fatal(err)
		}
		for _, tc := range cases {
			text := tc.expandedText()
			benchCount(b, "omnitoken", "count", encoding, tc.ID, text, func() int { return omni.CountTokens(text) })
			benchEncode(b, "omnitoken", "encode", encoding, tc.ID, text, func() []int { return omni.EncodeOrdinary(text) })
			omniTokens := omni.EncodeOrdinary(text)
			benchDecode(b, "omnitoken", "decode", encoding, tc.ID, text, func() string { return omni.Decode(omniTokens) })

			benchCount(b, "tiktoken_go", "count_by_encode", encoding, tc.ID, text, func() int { return len(goPort.EncodeOrdinary(text)) })
			benchEncode(b, "tiktoken_go", "encode", encoding, tc.ID, text, func() []int { return goPort.EncodeOrdinary(text) })
			goPortTokens := goPort.EncodeOrdinary(text)
			benchDecode(b, "tiktoken_go", "decode", encoding, tc.ID, text, func() string { return goPort.Decode(goPortTokens) })
		}
	}
}

func TestParityOpenAIBPE(t *testing.T) {
	for _, encoding := range []string{omnitoken.EncodingCL100KBase, omnitoken.EncodingO200KBase} {
		omni, err := omnitoken.ForEncoding(encoding)
		if err != nil {
			t.Fatal(err)
		}
		goPort, err := tiktoken.GetEncoding(encoding)
		if err != nil {
			t.Fatal(err)
		}
		for _, tc := range loadCorpus(t) {
			text := tc.expandedText()
			omniTokens := omni.EncodeOrdinary(text)
			goPortTokens := goPort.EncodeOrdinary(text)
			if !reflect.DeepEqual(omniTokens, goPortTokens) {
				t.Fatalf("%s/%s token mismatch\nomni=%v\ngoport=%v", encoding, tc.ID, omniTokens, goPortTokens)
			}
			if got := omni.Decode(omniTokens); got != text {
				t.Fatalf("%s/%s OmniToken decode mismatch", encoding, tc.ID)
			}
		}
	}
}

func benchCount(b *testing.B, runner, op, encoding, id, text string, fn func() int) {
	b.Run("runner="+runner+"/op="+op+"/enc="+encoding+"/case="+id, func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len([]byte(text))))
		for i := 0; i < b.N; i++ {
			countSink = fn()
		}
	})
}

func benchEncode(b *testing.B, runner, op, encoding, id, text string, fn func() []int) {
	b.Run("runner="+runner+"/op="+op+"/enc="+encoding+"/case="+id, func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len([]byte(text))))
		for i := 0; i < b.N; i++ {
			tokenSink = fn()
		}
	})
}

func benchDecode(b *testing.B, runner, op, encoding, id, text string, fn func() string) {
	b.Run("runner="+runner+"/op="+op+"/enc="+encoding+"/case="+id, func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len([]byte(text))))
		for i := 0; i < b.N; i++ {
			textSink = fn()
		}
	})
}

func loadCorpus(tb testing.TB) []corpusCase {
	tb.Helper()
	path := os.Getenv("OMNITOKEN_BENCH_CORPUS")
	if path == "" {
		path = filepath.Join("..", "..", "benchmarks", "corpus", "openai_bpe.jsonl")
	}
	file, err := os.Open(path)
	if err != nil {
		tb.Fatal(err)
	}
	defer file.Close()

	var cases []corpusCase
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var tc corpusCase
		if err := json.Unmarshal([]byte(line), &tc); err != nil {
			tb.Fatal(err)
		}
		cases = append(cases, tc)
	}
	if err := scanner.Err(); err != nil {
		tb.Fatal(err)
	}
	if len(cases) == 0 {
		tb.Fatal("empty benchmark corpus")
	}
	return cases
}

func (tc corpusCase) expandedText() string {
	if tc.Repeat <= 1 {
		return tc.Text
	}
	return strings.Repeat(tc.Text, tc.Repeat)
}
