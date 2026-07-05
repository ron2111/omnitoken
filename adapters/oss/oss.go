// Package oss provides optional local tokenizer adapters for user-supplied OSS model files.
package oss

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
	"time"

	sentencepiece "github.com/eliben/go-sentencepiece"
	omnitoken "github.com/ron2111/omnitoken"
)

// ModelSource points to a local or remote SentencePiece model.
type ModelSource struct {
	Data   []byte
	Path   string
	URL    string
	SHA256 string
}

// Options configures model loading.
type Options struct {
	Source     ModelSource
	CacheDir   string
	Offline    bool
	HTTPClient *http.Client
}

// EncodeOptions controls sequence-level special-token insertion.
type EncodeOptions struct {
	BOS bool
	EOS bool
}

// ModelInfo exposes selected SentencePiece metadata.
type ModelInfo struct {
	VocabularySize int
	BOSID          int
	EOSID          int
	UnknownID      int
	PadID          int
}

// Engine wraps a local SentencePiece processor as an OmniToken engine.
type Engine struct {
	name string
	proc *sentencepiece.Processor
	info ModelInfo
}

// NewSentencePiece builds an engine from a user-supplied SentencePiece model.
func NewSentencePiece(name string, opts Options) (*Engine, error) {
	if name == "" {
		return nil, errors.New("omnitoken oss: encoding name is required")
	}
	data, err := modelBytes(opts, opts.Source)
	if err != nil {
		return nil, err
	}
	proc, err := sentencepiece.NewProcessor(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	info := proc.ModelInfo()
	return &Engine{name: name, proc: proc, info: ModelInfo{
		VocabularySize: info.VocabularySize,
		BOSID:          info.BeginningOfSentenceID,
		EOSID:          info.EndOfSentenceID,
		UnknownID:      info.UnknownID,
		PadID:          info.PadID,
	}}, nil
}

// RegisterSentencePiece registers a local SentencePiece encoding.
func RegisterSentencePiece(encoding string, opts Options) error {
	return omnitoken.RegisterEncoding(encoding, func() (omnitoken.ModelEngine, error) {
		return NewSentencePiece(encoding, opts)
	})
}

// RegisterModels maps exact model names to an encoding.
func RegisterModels(provider omnitoken.Provider, encoding string, models ...string) error {
	for _, model := range models {
		if err := omnitoken.RegisterProviderModel(provider, model, encoding); err != nil {
			return err
		}
	}
	return nil
}

// RegisterModelPrefixes maps model-name prefixes to an encoding.
func RegisterModelPrefixes(provider omnitoken.Provider, encoding string, prefixes ...string) error {
	for _, prefix := range prefixes {
		if err := omnitoken.RegisterProviderModelPrefix(provider, prefix, encoding); err != nil {
			return err
		}
	}
	return nil
}

// Encoding returns the adapter encoding name.
func (e *Engine) Encoding() string {
	if e == nil {
		return ""
	}
	return e.name
}

// EncodeOrdinary encodes text without adding BOS/EOS tokens.
func (e *Engine) EncodeOrdinary(text string) []int {
	return e.Encode(text, EncodeOptions{})
}

// Encode encodes text and optionally inserts BOS/EOS IDs when the model defines them.
func (e *Engine) Encode(text string, opts EncodeOptions) []int {
	if e == nil || e.proc == nil || text == "" {
		return nil
	}
	tokens := e.proc.Encode(text)
	ids := make([]int, 0, len(tokens)+2)
	if opts.BOS && e.info.BOSID >= 0 {
		ids = append(ids, e.info.BOSID)
	}
	for _, token := range tokens {
		ids = append(ids, token.ID)
	}
	if opts.EOS && e.info.EOSID >= 0 {
		ids = append(ids, e.info.EOSID)
	}
	return ids
}

// CountTokens returns the ordinary token count.
func (e *Engine) CountTokens(text string) int {
	return e.Count(text, EncodeOptions{})
}

// Count returns the token count with optional BOS/EOS accounting.
func (e *Engine) Count(text string, opts EncodeOptions) int {
	if e == nil || e.proc == nil || text == "" {
		return 0
	}
	count := len(e.proc.Encode(text))
	if opts.BOS && e.info.BOSID >= 0 {
		count++
	}
	if opts.EOS && e.info.EOSID >= 0 {
		count++
	}
	return count
}

// Decode decodes token IDs with the configured SentencePiece model.
func (e *Engine) Decode(tokens []int) string {
	if e == nil || e.proc == nil || len(tokens) == 0 {
		return ""
	}
	return e.proc.Decode(tokens)
}

// ModelInfo returns selected metadata for the loaded model.
func (e *Engine) ModelInfo() ModelInfo {
	if e == nil {
		return ModelInfo{BOSID: -1, EOSID: -1, UnknownID: -1, PadID: -1}
	}
	return e.info
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
	return nil, errors.New("omnitoken oss: SentencePiece model data, path, or URL is required")
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
		return nil, fmt.Errorf("omnitoken oss: cached model unavailable in offline mode: %s", cachePath)
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
	tmpPath := cachePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return nil, err
	}
	if err := os.Rename(tmpPath, cachePath); err != nil {
		_ = os.Remove(tmpPath)
		return nil, err
	}
	return data, nil
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
		return nil, fmt.Errorf("omnitoken oss: download %s returned %s", url, resp.Status)
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
		return nil, fmt.Errorf("omnitoken oss: model sha256 = %s, want %s", got, want)
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
		cacheDir = filepath.Join(cacheDir, "omnitoken", "oss")
	}
	key := source.SHA256
	if key == "" {
		sum := sha256.Sum256([]byte(source.URL))
		key = hex.EncodeToString(sum[:])
	}
	return filepath.Join(cacheDir, key+".model"), nil
}
