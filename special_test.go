package omnitoken

import (
	"errors"
	"reflect"
	"testing"
)

func TestEncodeSpecialPolicy(t *testing.T) {
	engine, err := ForEncoding(EncodingO200KHarmony)
	if err != nil {
		t.Fatal(err)
	}

	plain := "hello world"
	got, err := Encode(engine, plain, EncodeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if want := engine.EncodeOrdinary(plain); !reflect.DeepEqual(got, want) {
		t.Fatalf("Encode ordinary text = %v, want %v", got, want)
	}

	ordinary := engine.EncodeOrdinary("<|start|>")
	if reflect.DeepEqual(ordinary, []int{200006}) {
		t.Fatalf("EncodeOrdinary encoded special marker as special ID: %v", ordinary)
	}

	if _, err := Encode(engine, "<|start|>", EncodeOptions{}); !errors.Is(err, ErrDisallowedSpecial) {
		t.Fatalf("Encode disallowed special err = %v, want ErrDisallowedSpecial", err)
	}

	got, err = Encode(engine, "<|start|>", EncodeOptions{AllowAllSpecial: true})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int{200006}) {
		t.Fatalf("Encode allowed special = %v, want [200006]", got)
	}

	got, err = Encode(engine, "hello <|start|> world", EncodeOptions{AllowedSpecial: map[string]bool{"<|start|>": true}})
	if err != nil {
		t.Fatal(err)
	}
	if !containsToken(got, 200006) {
		t.Fatalf("Encode specific allowed special = %v, want special ID in middle", got)
	}

	if _, err := Encode(engine, "<|start|><|end|>", EncodeOptions{AllowedSpecial: map[string]bool{"<|start|>": true}}); !errors.Is(err, ErrDisallowedSpecial) {
		t.Fatalf("Encode partially allowed special err = %v, want ErrDisallowedSpecial", err)
	}
}

func containsToken(tokens []int, want int) bool {
	for _, token := range tokens {
		if token == want {
			return true
		}
	}
	return false
}

func TestHarmonySpecialAlias(t *testing.T) {
	engine, err := ForEncoding(EncodingO200KHarmony)
	if err != nil {
		t.Fatal(err)
	}
	harmony := engine.(*Engine)

	for _, marker := range []string{"<|endofprompt|>", "<|reserved_200018|>"} {
		id, ok := harmony.SpecialTokenID(marker)
		if !ok {
			t.Fatalf("SpecialTokenID(%q) missing", marker)
		}
		if id != 200018 {
			t.Fatalf("SpecialTokenID(%q) = %d, want 200018", marker, id)
		}
		got, err := Encode(engine, marker, EncodeOptions{AllowAllSpecial: true})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, []int{200018}) {
			t.Fatalf("Encode(%q) = %v, want [200018]", marker, got)
		}
	}

	if got := engine.Decode([]int{200018}); got != "<|endofprompt|>" {
		t.Fatalf("Decode([200018]) = %q, want <|endofprompt|>", got)
	}
}

func TestCountTokensWithOptions(t *testing.T) {
	engine, err := ForEncoding(EncodingO200KHarmony)
	if err != nil {
		t.Fatal(err)
	}
	text := "<|start|>assistant<|message|>hello<|end|>"
	opts := EncodeOptions{AllowAllSpecial: true}
	tokens, err := Encode(engine, text, opts)
	if err != nil {
		t.Fatal(err)
	}
	count, err := CountTokensWithOptions(engine, text, opts)
	if err != nil {
		t.Fatal(err)
	}
	if count != len(tokens) {
		t.Fatalf("CountTokensWithOptions = %d, want len(Encode) %d (%v)", count, len(tokens), tokens)
	}
}
