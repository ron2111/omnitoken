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
	if _, ok := specials["<|reserved_special_token_0|>"]; ok {
		t.Fatal("reserved token 0 collides with begin_of_text")
	}
	if _, ok := specials["<|reserved_special_token_9|>"]; ok {
		t.Fatal("reserved token 9 collides with eot_id")
	}
	if specials["<|reserved_special_token_11|>"] != 128011 {
		t.Fatalf("reserved token 11 = %d", specials["<|reserved_special_token_11|>"])
	}
}

func TestNewRequiresSource(t *testing.T) {
	if _, err := New(Options{}); err == nil {
		t.Fatal("New accepted missing source")
	}
}
