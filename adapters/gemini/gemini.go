// Package gemini provides optional local Gemini text-tokenizer registration.
//
// The adapter uses local Gemma SentencePiece model files supplied by the user.
// It is intended for local text-tokenization estimates, not provider billing or
// multimodal accounting parity.
package gemini

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	omnitoken "github.com/ron2111/omnitoken"
	"github.com/ron2111/omnitoken/adapters/gemini/internal/gemmabpe"
)

const (
	EncodingGemma2 = "google_gemma2_sentencepiece"
	EncodingGemma3 = "google_gemma3_sentencepiece"

	gemma2URL    = "https://raw.githubusercontent.com/google/gemma_pytorch/33b652c465537c6158f9a472ea5700e5e770ad3f/tokenizer/tokenizer.model"
	gemma2SHA256 = "61a7b147390c64585d6c3543dd6fc636906c9af3865a5548f27f31aee1d4c8e2"
	gemma3URL    = "https://raw.githubusercontent.com/google/gemma_pytorch/014acb7ac4563a5f77c76d7ff98f31b568c16508/tokenizer/gemma3_cleaned_262144_v2.spiece.model"
	gemma3SHA256 = "1299c11d7cf632ef3b4e11937501358ada021bbdf7c47638d13c0ee982f2e79c"
)

// ModelSource points to a local SentencePiece model.
type ModelSource struct {
	Data   []byte
	Path   string
	URL    string
	SHA256 string
}

// Options configures local Gemini tokenizer registration.
type Options struct {
	Gemma2     ModelSource
	Gemma3     ModelSource
	CacheDir   string
	Offline    bool
	HTTPClient *http.Client
}

// Engine wraps a local SentencePiece processor as an OmniToken engine.
type Engine struct {
	name string
	proc *gemmabpe.Processor
}

var registration struct {
	sync.Mutex
	registered bool
	options    Options
}

// DefaultOptions returns official Google local-tokenizer sources with local caching enabled.
func DefaultOptions() Options {
	return Options{
		Gemma2: ModelSource{URL: gemma2URL, SHA256: gemma2SHA256},
		Gemma3: ModelSource{URL: gemma3URL, SHA256: gemma3SHA256},
	}
}

// Register registers Gemini model mappings using official local-tokenizer sources.
func Register() error { return RegisterWithOptions(DefaultOptions()) }

// RegisterWithOptions registers Gemini model mappings backed by local model sources.
func RegisterWithOptions(opts Options) error {
	registration.Lock()
	defer registration.Unlock()
	if registration.registered {
		return nil
	}
	registration.options = cloneOptions(opts)

	if err := omnitoken.RegisterEncoding(EncodingGemma2, func() (omnitoken.ModelEngine, error) {
		return newEngine(EncodingGemma2, currentOptions(), modelSource(EncodingGemma2))
	}); err != nil {
		return err
	}
	if err := omnitoken.RegisterEncoding(EncodingGemma3, func() (omnitoken.ModelEngine, error) {
		return newEngine(EncodingGemma3, currentOptions(), modelSource(EncodingGemma3))
	}); err != nil {
		return err
	}

	for _, model := range gemma2Models {
		if err := omnitoken.RegisterProviderModel(omnitoken.ProviderGoogle, model, EncodingGemma2); err != nil {
			return err
		}
	}
	for _, model := range gemma3Models {
		if err := omnitoken.RegisterProviderModel(omnitoken.ProviderGoogle, model, EncodingGemma3); err != nil {
			return err
		}
	}
	registration.registered = true
	return nil
}

func cloneOptions(opts Options) Options {
	opts.Gemma2.Data = cloneBytes(opts.Gemma2.Data)
	opts.Gemma3.Data = cloneBytes(opts.Gemma3.Data)
	return opts
}

func cloneBytes(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	return append([]byte(nil), data...)
}

// SupportedModels returns exact Gemini model mappings supported by the adapter.
func SupportedModels() []omnitoken.ModelInfo {
	models := make([]omnitoken.ModelInfo, 0, len(gemma2Models)+len(gemma3Models))
	for _, model := range gemma2Models {
		models = append(models, omnitoken.ModelInfo{Model: model, Provider: omnitoken.ProviderGoogle, Encoding: EncodingGemma2})
	}
	for _, model := range gemma3Models {
		models = append(models, omnitoken.ModelInfo{Model: model, Provider: omnitoken.ProviderGoogle, Encoding: EncodingGemma3})
	}
	return models
}

func currentOptions() Options {
	registration.Lock()
	defer registration.Unlock()
	return registration.options
}

func modelSource(encoding string) ModelSource {
	registration.Lock()
	defer registration.Unlock()
	if encoding == EncodingGemma2 {
		return registration.options.Gemma2
	}
	return registration.options.Gemma3
}

func newEngine(name string, opts Options, source ModelSource) (*Engine, error) {
	proc, err := newProcessor(opts, source)
	if err != nil {
		return nil, err
	}
	return &Engine{name: name, proc: proc}, nil
}

func newProcessor(opts Options, source ModelSource) (*gemmabpe.Processor, error) {
	data, err := modelBytes(opts, source)
	if err != nil {
		return nil, err
	}
	return gemmabpe.NewProcessor(bytes.NewReader(data))
}

func modelBytes(opts Options, source ModelSource) ([]byte, error) {
	if len(source.Data) > 0 {
		return verifiedBytes(source.Data, source.SHA256)
	}
	if source.Path != "" {
		data, err := os.ReadFile(source.Path)
		if err != nil {
			return nil, err
		}
		return verifiedBytes(data, source.SHA256)
	}
	if source.URL != "" {
		return cachedModelBytes(opts, source)
	}
	return nil, errors.New("omnitoken gemini: local SentencePiece model data, path, or URL is required")
}

func cachedModelBytes(opts Options, source ModelSource) ([]byte, error) {
	cachePath, err := cachePath(opts.CacheDir, source)
	if err != nil {
		return nil, err
	}
	if data, err := os.ReadFile(cachePath); err == nil {
		verified, err := verifiedBytes(data, source.SHA256)
		if err == nil {
			return verified, nil
		}
		if opts.Offline {
			return nil, err
		}
	}
	if opts.Offline {
		return nil, fmt.Errorf("omnitoken gemini: cached model unavailable in offline mode: %s", cachePath)
	}
	data, err := downloadModel(opts, source.URL)
	if err != nil {
		return nil, err
	}
	data, err = verifiedBytes(data, source.SHA256)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return nil, err
	}
	tmpPath := uniqueTempPath(cachePath)
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return nil, err
	}
	if err := os.Rename(tmpPath, cachePath); err != nil {
		_ = os.Remove(tmpPath)
		return nil, err
	}
	return data, nil
}

func uniqueTempPath(path string) string {
	return fmt.Sprintf("%s.%d.%d.tmp", path, os.Getpid(), time.Now().UnixNano())
}

func downloadModel(opts Options, url string) ([]byte, error) {
	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("omnitoken gemini: download %s returned %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func verifiedBytes(data []byte, want string) ([]byte, error) {
	if want == "" {
		return data, nil
	}
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if got != want {
		return nil, fmt.Errorf("omnitoken gemini: model sha256 = %s, want %s", got, want)
	}
	return data, nil
}

func cachePath(cacheDir string, source ModelSource) (string, error) {
	if cacheDir == "" {
		var err error
		cacheDir, err = os.UserCacheDir()
		if err != nil {
			cacheDir = os.TempDir()
		}
		cacheDir = filepath.Join(cacheDir, "omnitoken", "gemini")
	}
	key := source.SHA256
	if key == "" {
		sum := sha256.Sum256([]byte(source.URL))
		key = hex.EncodeToString(sum[:])
	}
	return filepath.Join(cacheDir, key+".model"), nil
}

// Encoding returns the adapter encoding name.
func (e *Engine) Encoding() string {
	if e == nil {
		return ""
	}
	return e.name
}

// EncodeOrdinary encodes text with the configured local SentencePiece model.
func (e *Engine) EncodeOrdinary(text string) []int {
	if e == nil || e.proc == nil || text == "" {
		return nil
	}
	return e.proc.EncodeIDs(text)
}

// CountTokens returns the number of local SentencePiece tokens in text.
func (e *Engine) CountTokens(text string) int {
	if e == nil || e.proc == nil || text == "" {
		return 0
	}
	return e.proc.CountTokens(text)
}

// Decode decodes token IDs with the configured local SentencePiece model.
func (e *Engine) Decode(tokens []int) string {
	if e == nil || e.proc == nil || len(tokens) == 0 {
		return ""
	}
	return e.proc.Decode(tokens)
}

var gemma2Models = []string{
	"gemini-1.0-pro",
	"gemini-1.0-pro-001",
	"gemini-1.0-pro-002",
	"gemini-1.5-pro",
	"gemini-1.5-pro-001",
	"gemini-1.5-pro-002",
	"gemini-1.5-flash",
	"gemini-1.5-flash-001",
	"gemini-1.5-flash-002",
}

var gemma3Models = []string{
	"gemini-2.0-flash",
	"gemini-2.0-flash-001",
	"gemini-2.0-flash-lite",
	"gemini-2.0-flash-lite-001",
	"gemini-2.5-pro",
	"gemini-2.5-pro-preview-06-05",
	"gemini-2.5-pro-preview-05-06",
	"gemini-2.5-pro-exp-03-25",
	"gemini-2.5-flash",
	"gemini-live-2.5-flash",
	"gemini-2.5-flash-preview-05-20",
	"gemini-2.5-flash-preview-04-17",
	"gemini-2.5-flash-lite",
	"gemini-2.5-flash-lite-preview-06-17",
	"gemini-3-pro-preview",
}
