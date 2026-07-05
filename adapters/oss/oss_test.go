package oss

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	omnitoken "github.com/ron2111/omnitoken"
)

func TestModelBytesValidation(t *testing.T) {
	data := []byte("model")
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

func TestRegisterModelsRequiresEncoding(t *testing.T) {
	if err := RegisterModels(omnitoken.ProviderMeta, "missing", "llama-test"); err == nil {
		t.Fatal("RegisterModels accepted missing encoding")
	}
	if err := RegisterModelPrefixes(ProviderMistral, "missing", "mistral-test-"); err == nil {
		t.Fatal("RegisterModelPrefixes accepted missing encoding")
	}
}
