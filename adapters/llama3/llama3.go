// Package llama3 provides optional local Llama 3 tiktoken-BPE support.
package llama3

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	omnitoken "github.com/ron2111/omnitoken"
)

const EncodingLlama3 = "meta_llama3_tiktoken_bpe"

// ModelSource points to a Llama 3 tiktoken-BPE tokenizer.model file.
type ModelSource struct {
	Data   []byte
	Path   string
	URL    string
	SHA256 string
}

// Options configures tokenizer loading and special-token layout.
type Options struct {
	Source     ModelSource
	CacheDir   string
	Offline    bool
	HTTPClient *http.Client
	Variant    Variant
}

// Variant controls known Llama 3 special-token names.
type Variant string

const (
	VariantLlama3  Variant = "llama3"
	VariantLlama31 Variant = "llama3.1"
	VariantLlama32 Variant = "llama3.2"
)

// EncodeOptions controls sequence-level special-token insertion.
type EncodeOptions struct {
	BOS bool
	EOS bool
}

// Engine wraps OmniToken's byte-BPE engine with Llama 3 metadata.
type Engine struct {
	inner    *omnitoken.Engine
	specials map[string]int
}

// New builds a Llama 3 adapter from a user-supplied tokenizer file.
func New(opts Options) (*Engine, error) {
	data, err := modelBytes(opts, opts.Source)
	if err != nil {
		return nil, err
	}
	specials := specialTokens(opts.Variant, 128000)
	inner, err := omnitoken.NewByteBPE(omnitoken.ByteBPEOptions{
		Name:      EncodingLlama3,
		Data:      data,
		Segmenter: omnitoken.SegmenterCL100K,
		Specials:  specials,
	})
	if err != nil {
		return nil, err
	}
	return &Engine{inner: inner, specials: specials}, nil
}

// Register registers the default Llama 3 encoding factory.
func Register(opts Options) error {
	return omnitoken.RegisterEncoding(EncodingLlama3, func() (omnitoken.ModelEngine, error) {
		return New(opts)
	})
}

// RegisterModelPrefixes maps model prefixes to the Llama 3 encoding.
func RegisterModelPrefixes(prefixes ...string) error {
	for _, prefix := range prefixes {
		if err := omnitoken.RegisterProviderModelPrefix(omnitoken.ProviderMeta, prefix, EncodingLlama3); err != nil {
			return err
		}
	}
	return nil
}

// Encoding returns the adapter encoding name.
func (e *Engine) Encoding() string {
	if e == nil || e.inner == nil {
		return ""
	}
	return e.inner.Encoding()
}

// EncodeOrdinary encodes text without adding BOS/EOS.
func (e *Engine) EncodeOrdinary(text string) []int {
	if e == nil || e.inner == nil {
		return nil
	}
	return e.inner.EncodeOrdinary(text)
}

// CountTokens returns the ordinary token count.
func (e *Engine) CountTokens(text string) int {
	if e == nil || e.inner == nil {
		return 0
	}
	return e.inner.CountTokens(text)
}

// Decode decodes token IDs.
func (e *Engine) Decode(tokens []int) string {
	if e == nil || e.inner == nil {
		return ""
	}
	return e.inner.Decode(tokens)
}

// Encode encodes text and optionally inserts BOS/EOS IDs.
func (e *Engine) Encode(text string, opts EncodeOptions) []int {
	ids := e.EncodeOrdinary(text)
	if opts.BOS {
		ids = append([]int{e.specials["<|begin_of_text|>"]}, ids...)
	}
	if opts.EOS {
		ids = append(ids, e.specials["<|end_of_text|>"])
	}
	return ids
}

// SpecialTokens returns a copy of the configured special-token table.
func (e *Engine) SpecialTokens() map[string]int {
	out := make(map[string]int, len(e.specials))
	for token, id := range e.specials {
		out[token] = id
	}
	return out
}

func specialTokens(variant Variant, base int) map[string]int {
	specials := make(map[string]int, 256)
	for i := 0; i < 256; i++ {
		specials[fmt.Sprintf("<|reserved_special_token_%d|>", i)] = base + i
	}
	specials["<|begin_of_text|>"] = base
	specials["<|end_of_text|>"] = base + 1
	specials["<|start_header_id|>"] = base + 6
	specials["<|end_header_id|>"] = base + 7
	specials["<|eot_id|>"] = base + 9
	if variant == VariantLlama31 || variant == VariantLlama32 {
		specials["<|finetune_right_pad_id|>"] = base + 4
		specials["<|eom_id|>"] = base + 8
		specials["<|python_tag|>"] = base + 10
	}
	return specials
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
	return nil, errors.New("omnitoken llama3: tokenizer data, path, or URL is required")
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
		return nil, fmt.Errorf("omnitoken llama3: cached tokenizer unavailable in offline mode: %s", cachePath)
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
		return nil, fmt.Errorf("omnitoken llama3: download %s returned %s", url, resp.Status)
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
		return nil, fmt.Errorf("omnitoken llama3: tokenizer sha256 = %s, want %s", got, want)
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
		cacheDir = filepath.Join(cacheDir, "omnitoken", "llama3")
	}
	key := source.SHA256
	if key == "" {
		sum := sha256.Sum256([]byte(source.URL))
		key = hex.EncodeToString(sum[:])
	}
	return filepath.Join(cacheDir, key+".model"), nil
}
