package omnitoken

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// WordPieceOptions configures a WordPiece tokenizer loaded from a newline vocab.
type WordPieceOptions struct {
	Name               string
	UnknownToken       string
	ContinuationPrefix string
	Lowercase          bool
}

// WordPieceEngine is a dependency-free WordPiece tokenizer engine.
type WordPieceEngine struct {
	name               string
	vocab              map[string]int
	decoder            map[int]string
	unknownID          int
	continuationPrefix string
	lowercase          bool
}

// NewWordPiece builds a WordPiece engine from a newline-delimited vocabulary.
func NewWordPiece(vocabData []byte, opts WordPieceOptions) (*WordPieceEngine, error) {
	name := opts.Name
	if name == "" {
		name = "wordpiece"
	}
	unknown := opts.UnknownToken
	if unknown == "" {
		unknown = "[UNK]"
	}
	prefix := opts.ContinuationPrefix
	if prefix == "" {
		prefix = "##"
	}

	vocab, decoder, err := parseWordPieceVocab(vocabData)
	if err != nil {
		return nil, err
	}
	unknownID, ok := vocab[unknown]
	if !ok {
		return nil, fmt.Errorf("omnitoken: wordpiece unknown token %q is missing", unknown)
	}

	return &WordPieceEngine{
		name:               name,
		vocab:              vocab,
		decoder:            decoder,
		unknownID:          unknownID,
		continuationPrefix: prefix,
		lowercase:          opts.Lowercase,
	}, nil
}

// Encoding returns the configured WordPiece engine name.
func (e *WordPieceEngine) Encoding() string {
	if e == nil {
		return ""
	}
	return e.name
}

// EncodeOrdinary encodes text using greedy longest-match WordPiece tokenization.
func (e *WordPieceEngine) EncodeOrdinary(text string) []int {
	if e == nil || text == "" {
		return nil
	}
	parts := wordPieceParts(text, e.lowercase)
	tokens := make([]int, 0, len(parts))
	for _, part := range parts {
		tokens = e.appendWordPiece(tokens, part)
	}
	return tokens
}

// CountTokens returns the number of WordPiece tokens in text.
func (e *WordPieceEngine) CountTokens(text string) int {
	if e == nil || text == "" {
		return 0
	}
	count := 0
	for _, part := range wordPieceParts(text, e.lowercase) {
		count += e.countWordPiece(part)
	}
	return count
}

// Decode decodes WordPiece IDs into normalized text.
func (e *WordPieceEngine) Decode(tokens []int) string {
	if e == nil || len(tokens) == 0 {
		return ""
	}
	var out strings.Builder
	for _, token := range tokens {
		piece, ok := e.decoder[token]
		if !ok {
			continue
		}
		if strings.HasPrefix(piece, e.continuationPrefix) {
			out.WriteString(strings.TrimPrefix(piece, e.continuationPrefix))
			continue
		}
		if out.Len() > 0 && !isWordPiecePunctuation(piece) {
			out.WriteByte(' ')
		}
		out.WriteString(piece)
	}
	return out.String()
}

func parseWordPieceVocab(data []byte) (map[string]int, map[int]string, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil, errors.New("omnitoken: empty wordpiece vocabulary")
	}
	vocab := make(map[string]int)
	decoder := make(map[int]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		piece := fields[0]
		if _, exists := vocab[piece]; exists {
			return nil, nil, fmt.Errorf("omnitoken: duplicate wordpiece token %q", piece)
		}
		id := len(vocab)
		vocab[piece] = id
		decoder[id] = piece
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	return vocab, decoder, nil
}

func (e *WordPieceEngine) appendWordPiece(dst []int, part string) []int {
	if id, ok := e.vocab[part]; ok {
		return append(dst, id)
	}
	runes := []rune(part)
	start := 0
	for start < len(runes) {
		bestID := -1
		bestEnd := start
		for end := len(runes); end > start; end-- {
			piece := string(runes[start:end])
			if start > 0 {
				piece = e.continuationPrefix + piece
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

func (e *WordPieceEngine) countWordPiece(part string) int {
	if _, ok := e.vocab[part]; ok {
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
				piece = e.continuationPrefix + piece
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

func wordPieceParts(text string, lowercase bool) []string {
	if lowercase {
		text = strings.ToLower(text)
	}
	parts := make([]string, 0, len(text)/4+1)
	start := -1
	for i, r := range text {
		if unicode.IsSpace(r) {
			if start >= 0 {
				parts = append(parts, text[start:i])
				start = -1
			}
			continue
		}
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
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

func isWordPiecePunctuation(piece string) bool {
	for _, r := range piece {
		return unicode.IsPunct(r) || unicode.IsSymbol(r)
	}
	return false
}
