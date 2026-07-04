package omnitoken

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
)

var testEncodingCounter uint64

const testWordPieceVocab = `
[UNK]
hello
world
un
##aff
##able
,
!
test
##ing
token
##izer
`

func TestWordPieceEncodeCountDecode(t *testing.T) {
	engine, err := NewWordPiece([]byte(testWordPieceVocab), WordPieceOptions{Lowercase: true})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		text string
		want []int
	}{
		{"Hello, world!", []int{1, 6, 2, 7}},
		{"unaffable", []int{3, 4, 5}},
		{"testing tokenizer", []int{8, 9, 10, 11}},
		{"missing", []int{0}},
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

	if got := engine.Decode([]int{1, 6, 2, 7}); got != "hello, world!" {
		t.Fatalf("Decode punctuation = %q", got)
	}
	if got := engine.Decode([]int{3, 4, 5}); got != "unaffable" {
		t.Fatalf("Decode continuation = %q", got)
	}
}

func TestWordPieceValidation(t *testing.T) {
	if _, err := NewWordPiece([]byte(""), WordPieceOptions{}); err == nil {
		t.Fatal("NewWordPiece empty vocab succeeded")
	}
	if _, err := NewWordPiece([]byte("hello\n"), WordPieceOptions{}); err == nil {
		t.Fatal("NewWordPiece without unknown token succeeded")
	}
	if _, err := NewWordPiece([]byte("[UNK]\n[UNK]\n"), WordPieceOptions{}); err == nil {
		t.Fatal("NewWordPiece duplicate vocab succeeded")
	}
}

func TestRegisterEncodingWithWordPiece(t *testing.T) {
	encoding := fmt.Sprintf("test_wordpiece_registry_%d", atomic.AddUint64(&testEncodingCounter, 1))
	if err := RegisterEncoding(encoding, func() (ModelEngine, error) {
		return NewWordPiece([]byte(testWordPieceVocab), WordPieceOptions{Name: encoding, Lowercase: true})
	}); err != nil {
		t.Fatal(err)
	}

	engine, err := ForEncoding(encoding)
	if err != nil {
		t.Fatal(err)
	}
	got := engine.EncodeOrdinary("Hello world")
	want := []int{1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("registered WordPiece EncodeOrdinary = %v, want %v", got, want)
	}
	if got := engine.(*WordPieceEngine).Encoding(); got != encoding {
		t.Fatalf("registered encoding name = %q, want %q", got, encoding)
	}
}

func TestRegisterEncodingRejectsInvalidInput(t *testing.T) {
	if err := RegisterEncoding("", func() (ModelEngine, error) { return nil, nil }); err == nil {
		t.Fatal("RegisterEncoding accepted empty name")
	}
	if err := RegisterEncoding("nil_factory", nil); err == nil {
		t.Fatal("RegisterEncoding accepted nil factory")
	}
	if err := RegisterEncoding(EncodingO200KBase, func() (ModelEngine, error) { return nil, nil }); err == nil {
		t.Fatal("RegisterEncoding allowed built-in override")
	}
}
