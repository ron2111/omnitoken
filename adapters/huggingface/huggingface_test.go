package huggingface

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"

	omnitoken "github.com/ron2111/omnitoken"
)

var testCounter uint64

const bertTokenizerJSON = `{
  "version": "1.0",
  "truncation": null,
  "padding": null,
  "added_tokens": [{"id": 10, "content": "[MASK]", "special": true}],
  "normalizer": {"type": "BertNormalizer", "lowercase": true, "strip_accents": true, "handle_chinese_chars": true},
  "pre_tokenizer": {"type": "BertPreTokenizer"},
  "decoder": {"type": "WordPiece"},
  "model": {
    "type": "WordPiece",
    "unk_token": "[UNK]",
    "continuing_subword_prefix": "##",
    "max_input_chars_per_word": 16,
    "vocab": {
      "[UNK]": 0,
      "hello": 1,
      ",": 2,
      "world": 3,
      "!": 4,
      "un": 5,
      "##aff": 6,
      "##able": 7,
      "中": 8,
      "文": 9,
      "[MASK]": 10
    }
  }
}`

func TestWordPieceTokenizerJSON(t *testing.T) {
	engine, err := NewTokenizerJSON([]byte(bertTokenizerJSON), Options{Name: "bert_test"})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		text string
		want []int
	}{
		{"Hello, world!", []int{1, 2, 3, 4}},
		{"unaffable", []int{5, 6, 7}},
		{"中文", []int{8, 9}},
		{"[MASK]", []int{10}},
		{"toolongword", []int{0}},
	}
	for _, tt := range tests {
		got := engine.EncodeOrdinary(tt.text)
		if !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("EncodeOrdinary(%q) = %v, want %v", tt.text, got, tt.want)
		}
		if count := engine.CountTokens(tt.text); count != len(tt.want) {
			t.Fatalf("CountTokens(%q) = %d, want %d", tt.text, count, len(tt.want))
		}
	}
	if got := engine.Decode([]int{1, 2, 3, 4}); got != "hello, world!" {
		t.Fatalf("Decode = %q", got)
	}
}

func TestRegisterTokenizerJSON(t *testing.T) {
	suffix := atomic.AddUint64(&testCounter, 1)
	encoding := fmt.Sprintf("hf_wordpiece_test_%d", suffix)
	model := fmt.Sprintf("hf-test-model-%d", suffix)
	if err := RegisterTokenizerJSON(encoding, []byte(bertTokenizerJSON), Options{}); err != nil {
		t.Fatal(err)
	}
	if err := omnitoken.RegisterModel(model, encoding); err != nil {
		t.Fatal(err)
	}
	engine, err := omnitoken.ForModel(model)
	if err != nil {
		t.Fatal(err)
	}
	if got := engine.CountTokens("Hello world"); got != 2 {
		t.Fatalf("CountTokens = %d, want 2", got)
	}
}

const bpeTokenizerJSON = `{
  "version": "1.0",
  "truncation": null,
  "padding": null,
  "pre_tokenizer": {"type": "WhitespaceSplit"},
  "decoder": {"type": "BPE"},
  "model": {
    "type": "BPE",
    "unk_token": "[UNK]",
    "vocab": {
      "a": 0,
      "b": 1,
      "c": 2,
      "ab": 3,
      "abc": 4,
      "x": 5,
      "[UNK]": 6
    },
    "merges": ["a b", ["ab", "c"]]
  }
}`

func TestBPETokenizerJSON(t *testing.T) {
	engine, err := NewTokenizerJSON([]byte(bpeTokenizerJSON), Options{Name: "bpe_test"})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		text string
		want []int
	}{
		{"abc", []int{4}},
		{"ab c", []int{3, 2}},
		{"x", []int{5}},
		{"z", []int{6}},
	}
	for _, tt := range tests {
		got := engine.EncodeOrdinary(tt.text)
		if !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("EncodeOrdinary(%q) = %v, want %v", tt.text, got, tt.want)
		}
		if count := engine.CountTokens(tt.text); count != len(tt.want) {
			t.Fatalf("CountTokens(%q) = %d, want %d", tt.text, count, len(tt.want))
		}
	}
}

func TestTokenizerJSONInfersBPEFromMerges(t *testing.T) {
	data := []byte(`{"model":{"unk_token":"[UNK]","vocab":{"a":0,"b":1,"ab":2,"[UNK]":3},"merges":["a b"]}}`)
	engine, err := NewTokenizerJSON(data, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := engine.EncodeOrdinary("ab"); !reflect.DeepEqual(got, []int{2}) {
		t.Fatalf("EncodeOrdinary = %v", got)
	}
}

const byteLevelBPETokenizerJSON = `{
  "version": "1.0",
  "truncation": null,
  "padding": null,
  "pre_tokenizer": {"type": "ByteLevel", "add_prefix_space": true, "use_regex": true},
  "decoder": {"type": "ByteLevel"},
  "model": {
    "type": "BPE",
    "vocab": {
      "Ġ": 0,
      "h": 1,
      "e": 2,
      "l": 3,
      "o": 4,
      "Ġh": 5,
      "Ġhe": 6,
      "Ġhel": 7,
      "Ġhell": 8,
      "Ġhello": 9,
      "w": 10,
      "r": 11,
      "d": 12,
      "Ġw": 13,
      "Ġwo": 14,
      "Ġwor": 15,
      "Ġworl": 16,
      "Ġworld": 17,
      "!": 18,
      "Ċ": 19
    },
    "merges": [
      "Ġ h", "Ġh e", "Ġhe l", "Ġhel l", "Ġhell o",
      "Ġ w", "Ġw o", "Ġwo r", "Ġwor l", "Ġworl d"
    ]
  }
}`

func TestByteLevelBPETokenizerJSON(t *testing.T) {
	engine, err := NewTokenizerJSON([]byte(byteLevelBPETokenizerJSON), Options{Name: "bytelevel_bpe_test"})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		text       string
		want       []int
		decodeWant string
	}{
		{"hello world!", []int{9, 17, 18}, " hello world!"},
		{"hello\n", []int{9, 19}, " hello\n"},
	}
	for _, tt := range tests {
		got := engine.EncodeOrdinary(tt.text)
		if !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("EncodeOrdinary(%q) = %v, want %v", tt.text, got, tt.want)
		}
		if got := engine.CountTokens(tt.text); got != len(tt.want) {
			t.Fatalf("CountTokens(%q) = %d, want %d", tt.text, got, len(tt.want))
		}
		if got := engine.Decode(tt.want); got != tt.decodeWant {
			t.Fatalf("Decode(%v) = %q, want %q", tt.want, got, tt.decodeWant)
		}
	}
}

func TestTokenizerJSONRejectsUnsupported(t *testing.T) {
	if _, err := NewTokenizerJSON([]byte(`{"model":{"type":"Unigram","vocab":[]}}`), Options{}); err == nil {
		t.Fatal("accepted unsupported Unigram model")
	}
	if _, err := NewTokenizerJSON([]byte(`{"truncation":{"max_length":4},"model":{"type":"WordPiece","unk_token":"[UNK]","vocab":{"[UNK]":0}}}`), Options{}); err == nil {
		t.Fatal("accepted truncation in strict mode")
	}
	if _, err := NewTokenizerJSON([]byte(`{"pre_tokenizer":{"type":"Sequence"},"model":{"type":"BPE","unk_token":"[UNK]","vocab":{"a":0,"[UNK]":1},"merges":["a a"]}}`), Options{}); err == nil {
		t.Fatal("accepted unsupported Sequence pre-tokenizer in strict mode")
	}
}
