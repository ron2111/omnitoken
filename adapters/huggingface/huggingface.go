// Package huggingface provides optional Hugging Face tokenizer.json adapters.
package huggingface

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	omnitoken "github.com/ron2111/omnitoken"
)

// Options configures tokenizer.json loading.
type Options struct {
	Name       string
	Permissive bool
}

// Engine implements a BERT-style WordPiece tokenizer.json subset.
type Engine struct {
	name         string
	vocab        map[string]int
	decoder      map[int]string
	unknownID    int
	unknownToken string
	prefix       string
	maxChars     int
	lowercase    bool
	stripAccents bool
	handleCJK    bool
	addedTokens  []string
}

type tokenizerJSON struct {
	Version      string          `json:"version"`
	Model        wordPieceModel  `json:"model"`
	Normalizer   component       `json:"normalizer"`
	PreTokenizer component       `json:"pre_tokenizer"`
	Decoder      component       `json:"decoder"`
	Truncation   json.RawMessage `json:"truncation"`
	Padding      json.RawMessage `json:"padding"`
	AddedTokens  []addedToken    `json:"added_tokens"`
}

type component struct {
	Type               string `json:"type"`
	Lowercase          bool   `json:"lowercase"`
	StripAccents       *bool  `json:"strip_accents"`
	HandleChineseChars *bool  `json:"handle_chinese_chars"`
}

type wordPieceModel struct {
	Type                    string         `json:"type"`
	UnkToken                string         `json:"unk_token"`
	ContinuingSubwordPrefix string         `json:"continuing_subword_prefix"`
	MaxInputCharsPerWord    int            `json:"max_input_chars_per_word"`
	Vocab                   map[string]int `json:"vocab"`
}

type addedToken struct {
	ID         int    `json:"id"`
	Content    string `json:"content"`
	Special    bool   `json:"special"`
	SingleWord bool   `json:"single_word"`
	LStrip     bool   `json:"lstrip"`
	RStrip     bool   `json:"rstrip"`
	Normalized bool   `json:"normalized"`
}

// NewTokenizerJSON builds an engine from a Hugging Face tokenizer.json file.
func NewTokenizerJSON(data []byte, opts Options) (*Engine, error) {
	var cfg tokenizerJSON
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Model.Type != "" && cfg.Model.Type != "WordPiece" {
		return nil, fmt.Errorf("omnitoken huggingface: unsupported model type %q", cfg.Model.Type)
	}
	if len(cfg.Model.Vocab) == 0 {
		return nil, errors.New("omnitoken huggingface: WordPiece vocab is required")
	}
	if !opts.Permissive {
		if len(cfg.Truncation) > 0 && string(cfg.Truncation) != "null" {
			return nil, errors.New("omnitoken huggingface: truncation is not supported by ModelEngine")
		}
		if len(cfg.Padding) > 0 && string(cfg.Padding) != "null" {
			return nil, errors.New("omnitoken huggingface: padding is not supported by ModelEngine")
		}
	}

	unknown := cfg.Model.UnkToken
	if unknown == "" {
		unknown = "[UNK]"
	}
	unknownID, ok := cfg.Model.Vocab[unknown]
	if !ok {
		return nil, fmt.Errorf("omnitoken huggingface: unknown token %q is missing", unknown)
	}
	prefix := cfg.Model.ContinuingSubwordPrefix
	if prefix == "" {
		prefix = "##"
	}
	maxChars := cfg.Model.MaxInputCharsPerWord
	if maxChars <= 0 {
		maxChars = 100
	}

	if err := validateComponents(cfg, opts.Permissive); err != nil {
		return nil, err
	}
	decoder := make(map[int]string, len(cfg.Model.Vocab))
	for piece, id := range cfg.Model.Vocab {
		if id < 0 {
			return nil, fmt.Errorf("omnitoken huggingface: negative vocab id for %q", piece)
		}
		if _, exists := decoder[id]; exists {
			return nil, fmt.Errorf("omnitoken huggingface: duplicate vocab id %d", id)
		}
		decoder[id] = piece
	}

	stripAccents := false
	if cfg.Normalizer.StripAccents != nil {
		stripAccents = *cfg.Normalizer.StripAccents
	} else if cfg.Normalizer.Lowercase {
		stripAccents = true
	}
	handleCJK := true
	if cfg.Normalizer.HandleChineseChars != nil {
		handleCJK = *cfg.Normalizer.HandleChineseChars
	}
	added := simpleAddedTokens(cfg.AddedTokens)

	name := opts.Name
	if name == "" {
		name = "huggingface_wordpiece"
	}
	return &Engine{
		name:         name,
		vocab:        cfg.Model.Vocab,
		decoder:      decoder,
		unknownID:    unknownID,
		unknownToken: unknown,
		prefix:       prefix,
		maxChars:     maxChars,
		lowercase:    cfg.Normalizer.Lowercase,
		stripAccents: stripAccents,
		handleCJK:    handleCJK,
		addedTokens:  added,
	}, nil
}

// RegisterTokenizerJSON registers a tokenizer.json as an OmniToken encoding.
func RegisterTokenizerJSON(name string, data []byte, opts Options) error {
	if name == "" {
		return errors.New("omnitoken huggingface: encoding name is required")
	}
	opts.Name = name
	return omnitoken.RegisterEncoding(name, func() (omnitoken.ModelEngine, error) {
		return NewTokenizerJSON(data, opts)
	})
}

// Encoding returns the adapter encoding name.
func (e *Engine) Encoding() string {
	if e == nil {
		return ""
	}
	return e.name
}

// EncodeOrdinary encodes text without tokenizer post-processing templates.
func (e *Engine) EncodeOrdinary(text string) []int {
	if e == nil || text == "" {
		return nil
	}
	parts := e.parts(text)
	tokens := make([]int, 0, len(parts))
	for _, part := range parts {
		tokens = e.appendPart(tokens, part)
	}
	return tokens
}

// CountTokens returns the ordinary token count.
func (e *Engine) CountTokens(text string) int {
	if e == nil || text == "" {
		return 0
	}
	count := 0
	for _, part := range e.parts(text) {
		count += e.countPart(part)
	}
	return count
}

// Decode decodes WordPiece token IDs into normalized text.
func (e *Engine) Decode(tokens []int) string {
	if e == nil || len(tokens) == 0 {
		return ""
	}
	var out strings.Builder
	for _, id := range tokens {
		piece, ok := e.decoder[id]
		if !ok {
			continue
		}
		if strings.HasPrefix(piece, e.prefix) {
			out.WriteString(strings.TrimPrefix(piece, e.prefix))
			continue
		}
		if out.Len() > 0 && !isPunctuation(piece) {
			out.WriteByte(' ')
		}
		out.WriteString(piece)
	}
	return out.String()
}

func validateComponents(cfg tokenizerJSON, permissive bool) error {
	if !permissive {
		if cfg.Normalizer.Type != "" && cfg.Normalizer.Type != "BertNormalizer" {
			return fmt.Errorf("omnitoken huggingface: unsupported normalizer %q", cfg.Normalizer.Type)
		}
		if cfg.PreTokenizer.Type != "" && cfg.PreTokenizer.Type != "BertPreTokenizer" {
			return fmt.Errorf("omnitoken huggingface: unsupported pre_tokenizer %q", cfg.PreTokenizer.Type)
		}
		if cfg.Decoder.Type != "" && cfg.Decoder.Type != "WordPiece" {
			return fmt.Errorf("omnitoken huggingface: unsupported decoder %q", cfg.Decoder.Type)
		}
	}
	return nil
}

func simpleAddedTokens(tokens []addedToken) []string {
	added := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token.Content == "" || token.SingleWord || token.LStrip || token.RStrip || token.Normalized {
			continue
		}
		if token.Special {
			added = append(added, token.Content)
		}
	}
	sort.Slice(added, func(i, j int) bool { return len(added[i]) > len(added[j]) })
	return added
}

func (e *Engine) parts(text string) []string {
	parts := make([]string, 0, len(text)/4+1)
	for i := 0; i < len(text); {
		if token, ok := e.matchAdded(text[i:]); ok {
			parts = append(parts, token)
			i += len(token)
			continue
		}
		r, size := runeAt(text, i)
		if unicode.IsSpace(r) {
			i += size
			continue
		}
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			parts = append(parts, string(r))
			i += size
			continue
		}
		start := i
		for i < len(text) {
			if token, ok := e.matchAdded(text[i:]); ok {
				_ = token
				break
			}
			r, size = runeAt(text, i)
			if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
				break
			}
			i += size
		}
		word := e.normalizeWord(text[start:i])
		if e.handleCJK {
			parts = appendCJKParts(parts, word)
		} else if word != "" {
			parts = append(parts, word)
		}
	}
	return parts
}

func (e *Engine) matchAdded(text string) (string, bool) {
	for _, token := range e.addedTokens {
		if strings.HasPrefix(text, token) {
			return token, true
		}
	}
	return "", false
}

func (e *Engine) normalizeWord(word string) string {
	if e.lowercase {
		word = strings.ToLower(word)
	}
	if !e.stripAccents {
		return word
	}
	var out strings.Builder
	for _, r := range word {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

func appendCJKParts(parts []string, text string) []string {
	start := -1
	for i, r := range text {
		if isCJK(r) {
			if start >= 0 {
				parts = append(parts, text[start:i])
				start = -1
			}
			parts = append(parts, string(r))
			continue
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		parts = append(parts, text[start:])
	}
	return parts
}

func (e *Engine) appendPart(dst []int, part string) []int {
	if id, ok := e.vocab[part]; ok {
		return append(dst, id)
	}
	if len([]rune(part)) > e.maxChars {
		return append(dst, e.unknownID)
	}
	runes := []rune(part)
	start := 0
	for start < len(runes) {
		bestID := -1
		bestEnd := start
		for end := len(runes); end > start; end-- {
			piece := string(runes[start:end])
			if start > 0 {
				piece = e.prefix + piece
			}
			if id, ok := e.vocab[piece]; ok {
				bestID = id
				bestEnd = end
				break
			}
		}
		if bestID < 0 {
			return append(dst, e.unknownID)
		}
		dst = append(dst, bestID)
		start = bestEnd
	}
	return dst
}

func (e *Engine) countPart(part string) int {
	if _, ok := e.vocab[part]; ok {
		return 1
	}
	if len([]rune(part)) > e.maxChars {
		return 1
	}
	runes := []rune(part)
	count := 0
	start := 0
	for start < len(runes) {
		bestEnd := start
		for end := len(runes); end > start; end-- {
			piece := string(runes[start:end])
			if start > 0 {
				piece = e.prefix + piece
			}
			if _, ok := e.vocab[piece]; ok {
				bestEnd = end
				break
			}
		}
		if bestEnd == start {
			return count + 1
		}
		count++
		start = bestEnd
	}
	return count
}

func runeAt(text string, i int) (rune, int) {
	for j, r := range text[i:] {
		return r, len(text[i : i+j+len(string(r))])
	}
	return 0, 0
}

func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || (r >= 0x3400 && r <= 0x4DBF) || (r >= 0x20000 && r <= 0x2A6DF)
}

func isPunctuation(piece string) bool {
	for _, r := range piece {
		return unicode.IsPunct(r) || unicode.IsSymbol(r)
	}
	return false
}
