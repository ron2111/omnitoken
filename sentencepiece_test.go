package omnitoken

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
)

const testSentencePieceVocab = `
<unk>
▁hello
▁world
▁token
izer
▁こんにちは
世界
!
`

var sentencePieceTestCounter uint64

func TestSentencePieceEncodeCountDecode(t *testing.T) {
	engine, err := NewSentencePiece([]byte(testSentencePieceVocab), SentencePieceOptions{AddDummyPrefix: true})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		text string
		want []int
	}{
		{"hello world!", []int{1, 2, 7}},
		{"tokenizer", []int{3, 4}},
		{"こんにちは世界", []int{5, 6}},
		{"unknown", []int{0, 0, 0, 0, 0, 0, 0, 0}},
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

	if got := engine.Decode([]int{1, 2, 7}); got != "hello world!" {
		t.Fatalf("Decode sentence = %q", got)
	}
	if got := engine.Decode([]int{3, 4}); got != "tokenizer" {
		t.Fatalf("Decode continuation = %q", got)
	}
}

func TestSentencePieceValidation(t *testing.T) {
	if _, err := NewSentencePiece([]byte(""), SentencePieceOptions{}); err == nil {
		t.Fatal("NewSentencePiece empty vocab succeeded")
	}
	if _, err := NewSentencePiece([]byte("▁hello\n"), SentencePieceOptions{}); err == nil {
		t.Fatal("NewSentencePiece without unknown token succeeded")
	}
	if _, err := NewSentencePiece([]byte("<unk>\n<unk>\n"), SentencePieceOptions{}); err == nil {
		t.Fatal("NewSentencePiece duplicate vocab succeeded")
	}
}

func TestRegisterModelWithSentencePiece(t *testing.T) {
	encoding := fmt.Sprintf("test_sentencepiece_registry_%d", atomic.AddUint64(&sentencePieceTestCounter, 1))
	model := encoding + "_model"
	if err := RegisterEncoding(encoding, func() (ModelEngine, error) {
		return NewSentencePiece([]byte(testSentencePieceVocab), SentencePieceOptions{Name: encoding, AddDummyPrefix: true})
	}); err != nil {
		t.Fatal(err)
	}
	if err := RegisterModel(model, encoding); err != nil {
		t.Fatal(err)
	}

	engine, err := ForModel(model)
	if err != nil {
		t.Fatal(err)
	}
	got := engine.EncodeOrdinary("hello world!")
	want := []int{1, 2, 7}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ForModel SentencePiece EncodeOrdinary = %v, want %v", got, want)
	}
}

func TestResolveProviderModel(t *testing.T) {
	encoding := fmt.Sprintf("test_provider_registry_%d", atomic.AddUint64(&sentencePieceTestCounter, 1))
	model := encoding + "_model"
	if err := RegisterEncoding(encoding, func() (ModelEngine, error) {
		return NewSentencePiece([]byte(testSentencePieceVocab), SentencePieceOptions{Name: encoding, AddDummyPrefix: true})
	}); err != nil {
		t.Fatal(err)
	}
	if err := RegisterProviderModel(ProviderGoogle, model, encoding); err != nil {
		t.Fatal(err)
	}

	info, err := ResolveModel(model)
	if err != nil {
		t.Fatal(err)
	}
	if info.Provider != ProviderGoogle || info.Encoding != encoding || info.Model != model {
		t.Fatalf("ResolveModel = %+v", info)
	}
}

func TestRegisterModelPrefixWithSentencePiece(t *testing.T) {
	encoding := fmt.Sprintf("test_sentencepiece_prefix_%d", atomic.AddUint64(&sentencePieceTestCounter, 1))
	prefix := encoding + "-"
	if err := RegisterEncoding(encoding, func() (ModelEngine, error) {
		return NewSentencePiece([]byte(testSentencePieceVocab), SentencePieceOptions{Name: encoding, AddDummyPrefix: true})
	}); err != nil {
		t.Fatal(err)
	}
	if err := RegisterProviderModelPrefix(ProviderMeta, prefix, encoding); err != nil {
		t.Fatal(err)
	}

	info, err := ResolveModel(prefix + "small")
	if err != nil {
		t.Fatal(err)
	}
	if info.Provider != ProviderMeta || info.Encoding != encoding {
		t.Fatalf("ResolveModel prefix = %+v", info)
	}

	engine, err := ForModel(prefix + "small")
	if err != nil {
		t.Fatal(err)
	}
	if got := engine.CountTokens("tokenizer"); got != 2 {
		t.Fatalf("prefix registered CountTokens = %d, want 2", got)
	}
}

func TestRegisterModelRejectsInvalidInput(t *testing.T) {
	if err := RegisterModel("", EncodingO200KBase); err == nil {
		t.Fatal("RegisterModel accepted empty model")
	}
	if err := RegisterModel("missing-encoding-model", "missing_encoding"); err == nil {
		t.Fatal("RegisterModel accepted missing encoding")
	}
	if err := RegisterProviderModel("", "empty-provider-model", EncodingO200KBase); err == nil {
		t.Fatal("RegisterProviderModel accepted empty provider")
	}
	if err := RegisterModel("gpt-4o", EncodingO200KBase); err == nil {
		t.Fatal("RegisterModel allowed built-in model override")
	}
	if err := RegisterModelPrefix("", EncodingO200KBase); err == nil {
		t.Fatal("RegisterModelPrefix accepted empty prefix")
	}
	if err := RegisterModelPrefix("gpt-4o-", EncodingO200KBase); err == nil {
		t.Fatal("RegisterModelPrefix allowed built-in prefix override")
	}
	if err := RegisterProviderModelPrefix("", "empty-provider-prefix-", EncodingO200KBase); err == nil {
		t.Fatal("RegisterProviderModelPrefix accepted empty provider")
	}
}
