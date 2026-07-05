package gemini

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
		"gemini-1.5-flash":                    {omnitoken.ProviderGoogle, EncodingGemma2},
		"gemini-1.5-flash-002":                {omnitoken.ProviderGoogle, EncodingGemma2},
		"gemini-2.5-flash":                    {omnitoken.ProviderGoogle, EncodingGemma3},
		"gemini-2.5-pro-preview-06-05":        {omnitoken.ProviderGoogle, EncodingGemma3},
		"gemini-2.5-flash-lite-preview-06-17": {omnitoken.ProviderGoogle, EncodingGemma3},
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
	for _, model := range []string{"gemini-future-unknown", "gemini-1.5-flash-8b"} {
		_, err := omnitoken.ResolveModel(model)
		if !errors.Is(err, omnitoken.ErrUnsupportedModel) {
			t.Fatalf("ResolveModel(%q) unsupported err = %v", model, err)
		}
	}
}

func TestDefaultOptionsUseOfficialSources(t *testing.T) {
	opts := DefaultOptions()
	if opts.Gemma2.URL != gemma2URL || opts.Gemma2.SHA256 != gemma2SHA256 {
		t.Fatalf("Gemma2 source = %+v", opts.Gemma2)
	}
	if opts.Gemma3.URL != gemma3URL || opts.Gemma3.SHA256 != gemma3SHA256 {
		t.Fatalf("Gemma3 source = %+v", opts.Gemma3)
	}
}

func TestSupportedModels(t *testing.T) {
	models := SupportedModels()
	if len(models) != len(gemma2Models)+len(gemma3Models) {
		t.Fatalf("SupportedModels len = %d", len(models))
	}
	found := map[string]string{}
	for _, info := range models {
		found[info.Model] = info.Encoding
		if info.Provider != omnitoken.ProviderGoogle {
			t.Fatalf("SupportedModels provider = %q", info.Provider)
		}
	}
	if found["gemini-1.0-pro-002"] != EncodingGemma2 {
		t.Fatalf("gemini-1.0-pro-002 mapping = %q", found["gemini-1.0-pro-002"])
	}
	if found["gemini-3-pro-preview"] != EncodingGemma3 {
		t.Fatalf("gemini-3-pro-preview mapping = %q", found["gemini-3-pro-preview"])
	}
}

func TestModelBytesHashVerification(t *testing.T) {
	data := []byte("model-bytes")
	sum := sha256.Sum256(data)
	want := hex.EncodeToString(sum[:])
	got, err := modelBytes(Options{}, ModelSource{Data: data, SHA256: want})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Fatalf("modelBytes = %q", got)
	}
	if _, err := modelBytes(Options{}, ModelSource{Data: data, SHA256: strings.Repeat("0", 64)}); err == nil {
		t.Fatal("modelBytes accepted wrong hash")
	}
}

func TestModelBytesOfflineCacheMiss(t *testing.T) {
	_, err := modelBytes(Options{CacheDir: t.TempDir(), Offline: true}, ModelSource{URL: "https://example.invalid/model", SHA256: strings.Repeat("1", 64)})
	if err == nil || !strings.Contains(err.Error(), "offline mode") {
		t.Fatalf("offline miss err = %v", err)
	}
}

func TestModelBytesRedownloadsCorruptCache(t *testing.T) {
	data := []byte("fresh-model")
	sum := sha256.Sum256(data)
	want := hex.EncodeToString(sum[:])
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(data)
	}))
	defer server.Close()

	cacheDir := t.TempDir()
	source := ModelSource{URL: server.URL + "/model", SHA256: want}
	path, err := cachePath(cacheDir, source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("corrupt"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := modelBytes(Options{CacheDir: cacheDir}, source)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Fatalf("modelBytes = %q", got)
	}
}
