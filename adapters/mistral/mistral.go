// Package mistral provides optional local Mistral Tekken tokenizer support.
package mistral

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	omnitoken "github.com/ron2111/omnitoken"
)

const ProviderMistral omnitoken.Provider = "mistral"
const EncodingTekken = "mistral_tekken_bpe"

// ModelSource points to a Tekken JSON tokenizer file.
type ModelSource struct {
	Data   []byte
	Path   string
	URL    string
	SHA256 string
}

// Options configures Tekken loading.
type Options struct {
	Source     ModelSource
	CacheDir   string
	Offline    bool
	HTTPClient *http.Client
}

// Engine wraps OmniToken's byte-BPE engine with Tekken metadata.
type Engine struct {
	inner    *omnitoken.Engine
	specials map[string]int
}

type tekkenFile struct {
	Config struct {
		Pattern string `json:"pattern"`
	} `json:"config"`
	DefaultNumSpecialTokens int           `json:"default_num_special_tokens"`
	Vocab                   []tekkenToken `json:"vocab"`
	SpecialTokens           []tekkenToken `json:"special_tokens"`
}

type tekkenToken struct {
	Rank       int    `json:"rank"`
	TokenBytes string `json:"token_bytes"`
	TokenStr   string `json:"token_str"`
	ID         int    `json:"id"`
}

// New builds a Mistral Tekken tokenizer from user-supplied JSON.
func New(opts Options) (*Engine, error) {
	data, err := modelBytes(opts, opts.Source)
	if err != nil {
		return nil, err
	}
	bpeData, specials, err := parseTekken(data)
	if err != nil {
		return nil, err
	}
	inner, err := omnitoken.NewByteBPE(omnitoken.ByteBPEOptions{
		Name:      EncodingTekken,
		Data:      bpeData,
		Segmenter: omnitoken.SegmenterCL100K,
		Specials:  specials,
	})
	if err != nil {
		return nil, err
	}
	return &Engine{inner: inner, specials: specials}, nil
}

// Register registers the Tekken encoding factory.
func Register(opts Options) error {
	return omnitoken.RegisterEncoding(EncodingTekken, func() (omnitoken.ModelEngine, error) {
		return New(opts)
	})
}

// RegisterModelPrefixes maps Mistral model-name prefixes to Tekken.
func RegisterModelPrefixes(prefixes ...string) error {
	for _, prefix := range prefixes {
		if err := omnitoken.RegisterProviderModelPrefix(ProviderMistral, prefix, EncodingTekken); err != nil {
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

func (e *Engine) EncodeOrdinary(text string) []int {
	if e == nil || e.inner == nil {
		return nil
	}
	return e.inner.EncodeOrdinary(text)
}

func (e *Engine) CountTokens(text string) int {
	if e == nil || e.inner == nil {
		return 0
	}
	return e.inner.CountTokens(text)
}

func (e *Engine) Decode(tokens []int) string {
	if e == nil || e.inner == nil {
		return ""
	}
	return e.inner.Decode(tokens)
}

// SpecialTokens returns configured special-token IDs.
func (e *Engine) SpecialTokens() map[string]int {
	out := make(map[string]int, len(e.specials))
	for token, id := range e.specials {
		out[token] = id
	}
	return out
}

func parseTekken(data []byte) ([]byte, map[string]int, error) {
	var file tekkenFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, nil, err
	}
	if len(file.Vocab) == 0 {
		return nil, nil, errors.New("omnitoken mistral: tekken vocab is required")
	}
	specialOffset := file.DefaultNumSpecialTokens
	if specialOffset < 0 {
		return nil, nil, errors.New("omnitoken mistral: negative special token offset")
	}
	sort.Slice(file.Vocab, func(i, j int) bool { return file.Vocab[i].Rank < file.Vocab[j].Rank })
	var out bytes.Buffer
	for _, token := range file.Vocab {
		if token.Rank < 0 {
			return nil, nil, fmt.Errorf("omnitoken mistral: negative rank %d", token.Rank)
		}
		if token.TokenBytes == "" {
			return nil, nil, errors.New("omnitoken mistral: token_bytes is required")
		}
		raw, err := base64.StdEncoding.DecodeString(token.TokenBytes)
		if err != nil {
			return nil, nil, err
		}
		_, _ = fmt.Fprintf(&out, "%s %d\n", base64.StdEncoding.EncodeToString(raw), token.Rank+specialOffset)
	}
	specials := make(map[string]int, len(file.SpecialTokens))
	for _, token := range file.SpecialTokens {
		name := token.TokenStr
		if name == "" && token.TokenBytes != "" {
			raw, err := base64.StdEncoding.DecodeString(token.TokenBytes)
			if err != nil {
				return nil, nil, err
			}
			name = string(raw)
		}
		if name == "" {
			continue
		}
		id := token.ID
		if id == 0 {
			id = token.Rank
		}
		specials[name] = id
	}
	return out.Bytes(), specials, nil
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
	return nil, errors.New("omnitoken mistral: Tekken data, path, or URL is required")
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
		return nil, fmt.Errorf("omnitoken mistral: cached tokenizer unavailable in offline mode: %s", cachePath)
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
		return nil, fmt.Errorf("omnitoken mistral: download %s returned %s", url, resp.Status)
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
		return nil, fmt.Errorf("omnitoken mistral: tokenizer sha256 = %s, want %s", got, want)
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
		cacheDir = filepath.Join(cacheDir, "omnitoken", "mistral")
	}
	key := source.SHA256
	if key == "" {
		sum := sha256.Sum256([]byte(source.URL))
		key = hex.EncodeToString(sum[:])
	}
	return filepath.Join(cacheDir, key+".json"), nil
}
