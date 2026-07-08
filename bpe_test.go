package omnitoken

import (
	"reflect"
	"testing"
)

func TestNewByteBPE(t *testing.T) {
	specials := map[string]int{"<|test|>": 100300}
	engine, err := NewByteBPE(ByteBPEOptions{
		Name:      "test_bpe",
		Data:      cl100kBaseData,
		Segmenter: SegmenterCL100K,
		Specials:  specials,
	})
	if err != nil {
		t.Fatal(err)
	}
	specials["<|test|>"] = 100301
	specials["<|mutated|>"] = 100302
	got := engine.EncodeOrdinary("hello world")
	want := []int{15339, 1917}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("EncodeOrdinary = %v, want %v", got, want)
	}
	if got := engine.Decode([]int{100300}); got != "<|test|>" {
		t.Fatalf("special decode = %q", got)
	}
	if _, ok := engine.SpecialTokenID("<|mutated|>"); ok {
		t.Fatal("engine retained mutable caller specials map")
	}
}

func TestNewByteBPEValidation(t *testing.T) {
	if _, err := NewByteBPE(ByteBPEOptions{}); err == nil {
		t.Fatal("NewByteBPE accepted empty options")
	}
	if _, err := NewByteBPE(ByteBPEOptions{Name: "x", Data: []byte("bad"), Segmenter: "missing"}); err == nil {
		t.Fatal("NewByteBPE accepted unsupported segmenter")
	}
}
