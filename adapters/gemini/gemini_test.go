package gemini

import (
	"errors"
	"testing"

	omnitoken "github.com/ron2111/omnitoken"
)

func TestRegisterWithOptionsMapsGeminiModels(t *testing.T) {
	if err := RegisterWithOptions(Options{}); err != nil {
		t.Fatal(err)
	}
	if err := RegisterWithOptions(Options{}); err != nil {
		t.Fatal(err)
	}

	tests := map[string]struct {
		provider omnitoken.Provider
		encoding string
	}{
		"gemini-1.5-flash": {omnitoken.ProviderGoogle, EncodingGemma2},
		"gemini-2.5-flash": {omnitoken.ProviderGoogle, EncodingGemma3},
	}
	for model, want := range tests {
		info, err := omnitoken.ResolveModel(model)
		if err != nil {
			t.Fatalf("ResolveModel(%q): %v", model, err)
		}
		if info.Provider != want.provider || info.Encoding != want.encoding {
			t.Fatalf("ResolveModel(%q) = %+v, want provider=%q encoding=%q", model, info, want.provider, want.encoding)
		}
	}
}

func TestGeminiEngineRequiresLocalModel(t *testing.T) {
	_, err := omnitoken.ForEncoding(EncodingGemma3)
	if err == nil {
		t.Fatal("ForEncoding without model source succeeded")
	}
}

func TestUnsupportedGeminiModelStaysUnsupported(t *testing.T) {
	_, err := omnitoken.ResolveModel("gemini-future-unknown")
	if !errors.Is(err, omnitoken.ErrUnsupportedModel) {
		t.Fatalf("ResolveModel unsupported err = %v", err)
	}
}
