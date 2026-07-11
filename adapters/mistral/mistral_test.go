package mistral

import "testing"

const tinyTekken = `{
  "default_num_special_tokens": 1000,
  "vocab": [
    {"rank": 0, "token_bytes": "YQ=="},
    {"rank": 1, "token_bytes": "Yg=="},
    {"rank": 2, "token_bytes": "YWI="}
  ],
  "special_tokens": [
    {"id": 0, "token_str": "<s>"},
    {"id": 2, "token_str": "</s>"}
  ]
}`

func TestParseTekken(t *testing.T) {
	data, specials, pattern, err := parseTekken([]byte(tinyTekken))
	if err != nil {
		t.Fatal(err)
	}
	if pattern != "" {
		t.Fatalf("pattern = %q", pattern)
	}
	if len(data) == 0 {
		t.Fatal("empty BPE data")
	}
	if specials["<s>"] != 0 || specials["</s>"] != 2 {
		t.Fatalf("specials = %+v", specials)
	}
}

func TestNewRequiresSource(t *testing.T) {
	if _, err := New(Options{}); err == nil {
		t.Fatal("New accepted missing source")
	}
}

func TestNewRejectsUnsupportedPatternByDefault(t *testing.T) {
	data := []byte(`{
  "config": {"pattern": "custom"},
  "default_num_special_tokens": 1000,
  "vocab": [{"rank": 0, "token_bytes": "YQ=="}]
}`)
	if _, err := New(Options{Source: ModelSource{Data: data}}); err == nil {
		t.Fatal("New accepted unsupported Tekken pattern")
	}
	if _, err := New(Options{Source: ModelSource{Data: data}, AllowUnsupportedPattern: true}); err != nil {
		t.Fatalf("New with AllowUnsupportedPattern: %v", err)
	}
}
