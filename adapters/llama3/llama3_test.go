package llama3

import "testing"

func TestSpecialTokens(t *testing.T) {
	specials := specialTokens(VariantLlama31, 128000)
	if specials["<|begin_of_text|>"] != 128000 || specials["<|eot_id|>"] != 128009 {
		t.Fatalf("specials = %+v", specials)
	}
	if specials["<|python_tag|>"] != 128010 {
		t.Fatalf("python tag = %d", specials["<|python_tag|>"])
	}
}

func TestNewRequiresSource(t *testing.T) {
	if _, err := New(Options{}); err == nil {
		t.Fatal("New accepted missing source")
	}
}
