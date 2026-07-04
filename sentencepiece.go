package omnitoken

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// SentencePieceOptions configures a lightweight SentencePiece-style tokenizer.
type SentencePieceOptions struct {
	Name           string
	UnknownToken   string
	Metaspace      string
	AddDummyPrefix bool
}

// SentencePieceEngine tokenizes text with a local metaspace vocabulary.
type SentencePieceEngine struct {
	name           string
	vocab          map[string]int
	decoder        map[int]string
	unknownID      int
	metaspace      string
	addDummyPrefix bool
}

// NewSentencePiece builds a SentencePiece-style engine from a newline vocabulary.
func NewSentencePiece(vocabData []byte, opts SentencePieceOptions) (*SentencePieceEngine, error) {
	name := opts.Name
	if name == "" {
		name = "sentencepiece"
	}
	unknown := opts.UnknownToken
	if unknown == "" {
		unknown = "<unk>"
	}
	metaspace := opts.Metaspace
	if metaspace == "" {
		metaspace = "▁"
	}

	vocab, decoder, err := parseSentencePieceVocab(vocabData)
	if err != nil {
		return nil, err
	}
	unknownID, ok := vocab[unknown]
	if !ok {
		return nil, fmt.Errorf("omnitoken: sentencepiece unknown token %q is missing", unknown)
	}

	return &SentencePieceEngine{
		name:           name,
		vocab:          vocab,
		decoder:        decoder,
		unknownID:      unknownID,
		metaspace:      metaspace,
		addDummyPrefix: opts.AddDummyPrefix,
	}, nil
}

// Encoding returns the configured SentencePiece engine name.
func (e *SentencePieceEngine) Encoding() string {
	if e == nil {
		return ""
	}
	return e.name
}

// EncodeOrdinary encodes text using greedy longest-match metaspace tokenization.
func (e *SentencePieceEngine) EncodeOrdinary(text string) []int {
	if e == nil || text == "" {
		return nil
	}
	normalized := e.normalize(text)
	tokens := make([]int, 0, len(normalized)/4+1)
	return e.appendPieces(tokens, normalized)
}

// CountTokens returns the number of SentencePiece-style tokens in text.
func (e *SentencePieceEngine) CountTokens(text string) int {
	if e == nil || text == "" {
		return 0
	}
	return e.countPieces(e.normalize(text))
}

// Decode decodes token IDs and converts metaspace markers back to spaces.
func (e *SentencePieceEngine) Decode(tokens []int) string {
	if e == nil || len(tokens) == 0 {
		return ""
	}
	var out strings.Builder
	for _, token := range tokens {
		piece, ok := e.decoder[token]
		if !ok {
			continue
		}
		out.WriteString(piece)
	}
	text := strings.ReplaceAll(out.String(), e.metaspace, " ")
	if e.addDummyPrefix {
		text = strings.TrimPrefix(text, " ")
	}
	return text
}

func parseSentencePieceVocab(data []byte) (map[string]int, map[int]string, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil, errors.New("omnitoken: empty sentencepiece vocabulary")
	}
	vocab := make(map[string]int)
	decoder := make(map[int]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		piece := strings.Fields(line)[0]
		if _, exists := vocab[piece]; exists {
			return nil, nil, fmt.Errorf("omnitoken: duplicate sentencepiece token %q", piece)
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

func (e *SentencePieceEngine) normalize(text string) string {
	if e.addDummyPrefix && !strings.HasPrefix(text, " ") {
		text = " " + text
	}
	return strings.ReplaceAll(text, " ", e.metaspace)
}

func (e *SentencePieceEngine) appendPieces(dst []int, text string) []int {
	if id, ok := e.vocab[text]; ok {
		return append(dst, id)
	}
	runes := []rune(text)
	for start := 0; start < len(runes); {
		bestID := -1
		bestEnd := start
		for end := len(runes); end > start; end-- {
			piece := string(runes[start:end])
			if id, ok := e.vocab[piece]; ok {
				bestID = id
				bestEnd = end
				break
			}
		}
		if bestID < 0 {
			dst = append(dst, e.unknownID)
			start++
			continue
		}
		dst = append(dst, bestID)
		start = bestEnd
	}
	return dst
}

func (e *SentencePieceEngine) countPieces(text string) int {
	if _, ok := e.vocab[text]; ok {
		return 1
	}
	runes := []rune(text)
	count := 0
	for start := 0; start < len(runes); {
		bestEnd := start
		for end := len(runes); end > start; end-- {
			piece := string(runes[start:end])
			if _, ok := e.vocab[piece]; ok {
				bestEnd = end
				break
			}
		}
		if bestEnd == start {
			start++
		} else {
			start = bestEnd
		}
		count++
	}
	return count
}
