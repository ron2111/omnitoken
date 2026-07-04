package omnitoken

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"
)

const referenceCorpusSize = 50000

var referenceCorpusDigest = map[string]string{
	EncodingCL100KBase: "bd667b8ad7459eb9cdc28ca4dae8a03cc7d334e99a40f1558948ec9c2fbea181",
	EncodingO200KBase:  "875400dbed51cb7c9a06f61308a1491f0adc062e25ed2a2397c397b0157d255a",
}

func TestOpenAIParityCorpusDigest(t *testing.T) {
	for _, encoding := range []string{EncodingCL100KBase, EncodingO200KBase} {
		t.Run(encoding, func(t *testing.T) {
			engine, err := ForEncoding(encoding)
			if err != nil {
				t.Fatal(err)
			}

			got := digestReferenceCorpus(t, engine, encoding)
			want := referenceCorpusDigest[encoding]
			if want == "" {
				t.Fatalf("missing %s corpus digest; run tools/openai_parity_digest.py with OpenAI tiktoken", encoding)
			}
			if got != want {
				t.Fatalf("%s corpus digest = %s, want %s", encoding, got, want)
			}
		})
	}
}

func TestO200KHarmonyCorpusMatchesO200KBase(t *testing.T) {
	base, err := ForEncoding(EncodingO200KBase)
	if err != nil {
		t.Fatal(err)
	}
	harmony, err := ForEncoding(EncodingO200KHarmony)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < referenceCorpusSize; i++ {
		text := referenceCorpusText(i)
		baseTokens := base.EncodeOrdinary(text)
		harmonyTokens := harmony.EncodeOrdinary(text)
		if fmt.Sprint(harmonyTokens) != fmt.Sprint(baseTokens) {
			t.Fatalf("o200k_harmony ordinary tokens differ at corpus index %d for %q\nbase=%v\nharmony=%v", i, text, baseTokens, harmonyTokens)
		}
	}
}

func digestReferenceCorpus(t *testing.T, engine ModelEngine, encoding string) string {
	t.Helper()

	hash := sha256.New()
	var buf [8]byte
	for i := 0; i < referenceCorpusSize; i++ {
		text := referenceCorpusText(i)
		tokens := engine.EncodeOrdinary(text)
		if count := engine.CountTokens(text); count != len(tokens) {
			t.Fatalf("%s count mismatch at corpus index %d: CountTokens=%d EncodeOrdinary len=%d text=%q", encoding, i, count, len(tokens), text)
		}
		if decoded := engine.Decode(tokens); decoded != text {
			t.Fatalf("%s decode mismatch at corpus index %d: got %q, want %q", encoding, i, decoded, text)
		}

		binary.LittleEndian.PutUint32(buf[:4], uint32(i))
		hash.Write(buf[:4])
		hash.Write([]byte(encoding))
		hash.Write([]byte{0})
		hash.Write([]byte(text))
		hash.Write([]byte{0})
		binary.LittleEndian.PutUint32(buf[:4], uint32(len(tokens)))
		hash.Write(buf[:4])
		for _, token := range tokens {
			binary.LittleEndian.PutUint32(buf[:4], uint32(token))
			hash.Write(buf[:4])
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func referenceCorpusText(i int) string {
	words := [...]string{"hello", "world", "token", "cache", "scanner", "BPE", "OpenAI", "gpt-4o", "JSON", "markdown", "unicode", "throughput"}
	cjk := [...]string{"こんにちは世界", "中文测试", "안녕하세요 세계", "ภาษาไทยทดสอบ", "مرحبا بالعالم"}
	emoji := [...]string{"😀", "🚀", "👩‍💻", "🔥", "✨", "🧪", "🌍", "✅"}
	code := [...]string{"func main() { return }", "if err != nil { return err }", "const value = items[index]", "SELECT * FROM users WHERE id = 123", "for i := 0; i < n; i++ { sum += i }"}
	markdown := [...]string{"# Title\n\n- item one\n- item two", "**bold** _italic_ `code`", "> quoted text\n\n```go\nfmt.Println(x)\n```", "[link](https://example.com/path?q=token)"}
	spaces := [...]string{" ", "  ", "\t", "\n", " \n", "\r\n", "   leading", "trailing   ", "middle   gap"}

	switch i % 16 {
	case 0:
		return fmt.Sprintf("%s %s %d", words[i%len(words)], words[(i*7+3)%len(words)], i)
	case 1:
		return fmt.Sprintf("I'm testing %d%d%d tokens, you're checking counts.", i%1000, (i*7)%1000, (i*13)%1000)
	case 2:
		return fmt.Sprintf("{\"id\":%d,\"name\":\"%s\",\"active\":%t,\"score\":%d}", i, words[i%len(words)], i%2 == 0, i*17)
	case 3:
		return markdown[i%len(markdown)] + fmt.Sprintf("\n\nParagraph %d with %s.", i, words[(i+5)%len(words)])
	case 4:
		return code[i%len(code)] + fmt.Sprintf(" // case_%d", i)
	case 5:
		return fmt.Sprintf("%s %s %s %d", cjk[i%len(cjk)], words[i%len(words)], emoji[i%len(emoji)], i)
	case 6:
		return fmt.Sprintf("snake_case/path-to/file_%d.go::FunctionName", i)
	case 7:
		return fmt.Sprintf("HTTPServerError%d ABCdefGHI XYZabc", i)
	case 8:
		return fmt.Sprintf("%s%s%s%s", spaces[i%len(spaces)], words[i%len(words)], spaces[(i+3)%len(spaces)], words[(i+4)%len(words)])
	case 9:
		return fmt.Sprintf("Numbers: %03d %06d %09d %.2f", i%1000, i*37%1000000, i*7919, float64(i)/7.0)
	case 10:
		return fmt.Sprintf("Symbols []{}()<> +=-*/ %% ^ & | ~ ! ? #%d", i)
	case 11:
		return fmt.Sprintf("URLs/email: https://example.com/%s/%d?a=b&c=d user%d@example.com", words[i%len(words)], i, i)
	case 12:
		return longReferencePrompt(i, words[:], emoji[:])
	case 13:
		return fmt.Sprintf("Mixed scripts %s %s %s caf\u00e9 e\u0301 na\u00efve", cjk[(i+1)%len(cjk)], emoji[(i+2)%len(emoji)], words[(i+3)%len(words)])
	case 14:
		return fmt.Sprintf("<|endoftext|> is ordinary text here %d <|start|><|channel|><|end|>", i)
	default:
		return fmt.Sprintf("line one %d\nline two\r\n\tindented %s %s", i, words[i%len(words)], emoji[i%len(emoji)])
	}
}

func longReferencePrompt(i int, words []string, emoji []string) string {
	text := "System: You are a tokenizer benchmark assistant."
	for j := 0; j < 20; j++ {
		text += fmt.Sprintf("\nStep %02d: preserve %s, count %d, emit %s safely.", j, words[(i+j)%len(words)], i*j+j, emoji[(i+j)%len(emoji)])
	}
	return text
}
