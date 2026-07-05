package omnitoken

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"
)

type segmenterFunc func(src []byte, start int) int

func (f segmenterFunc) Next(src []byte, start int) int { return f(src, start) }

// Engine is a loaded tokenizer encoding with mergeable ranks and decode tables.
type Engine struct {
	name      string
	segmenter Segmenter
	ranks     map[string]int
	decoder   [][]byte
	specials  map[string]int
}

// Encoding returns the OpenAI-compatible encoding name used by this engine.
func (e *Engine) Encoding() string { return e.name }

// EncodeOrdinary encodes text as ordinary text, without interpreting special tokens.
func (e *Engine) EncodeOrdinary(text string) []int {
	if e == nil || text == "" {
		return nil
	}

	src := unsafeStringBytes(text)
	tokens := make([]int, 0, len(src)/4+1)
	for start := 0; start < len(src); {
		end := e.segmenter.Next(src, start)
		if end <= start || end > len(src) {
			end = nextRuneIndex(src, start)
		}
		tokens = e.appendPieceTokens(tokens, src[start:end])
		start = end
	}
	return tokens
}

// CountTokens returns the number of ordinary tokens in text.
func (e *Engine) CountTokens(text string) int {
	if e == nil || text == "" {
		return 0
	}

	src := unsafeStringBytes(text)
	count := 0
	for start := 0; start < len(src); {
		end := e.segmenter.Next(src, start)
		if end <= start || end > len(src) {
			end = nextRuneIndex(src, start)
		}
		count += e.countPieceTokens(src[start:end])
		start = end
	}
	return count
}

// Decode decodes token IDs back to UTF-8 text. Unknown IDs are skipped.
func (e *Engine) Decode(tokens []int) string {
	if e == nil || len(tokens) == 0 {
		return ""
	}

	out := make([]byte, 0, len(tokens)*4)
	for _, token := range tokens {
		if token >= 0 && token < len(e.decoder) {
			raw := e.decoder[token]
			out = append(out, raw...)
		}
	}
	return string(out)
}

func (e *Engine) appendPieceTokens(dst []int, piece []byte) []int {
	if rank, ok := e.ranks[string(piece)]; ok {
		return append(dst, rank)
	}
	return e.bytePairEncode(dst, piece)
}

func (e *Engine) countPieceTokens(piece []byte) int {
	if _, ok := e.ranks[string(piece)]; ok {
		return 1
	}
	return e.bytePairCount(piece)
}

type bpePart struct {
	start int
	end   int
}

func (e *Engine) bytePairEncode(dst []int, piece []byte) []int {
	var stack [256]bpePart
	parts := initialBPEParts(piece, stack[:])
	parts = e.mergeBPEParts(piece, parts)
	for _, part := range parts {
		if rank, ok := e.ranks[string(piece[part.start:part.end])]; ok {
			dst = append(dst, rank)
		}
	}
	return dst
}

func (e *Engine) bytePairCount(piece []byte) int {
	var stack [256]bpePart
	parts := initialBPEParts(piece, stack[:])
	parts = e.mergeBPEParts(piece, parts)
	return len(parts)
}

func initialBPEParts(piece []byte, scratch []bpePart) []bpePart {
	parts := scratch
	if len(piece) > len(parts) {
		parts = make([]bpePart, len(piece))
	} else {
		parts = parts[:len(piece)]
	}
	for i := range piece {
		parts[i] = bpePart{start: i, end: i + 1}
	}
	return parts
}

func (e *Engine) mergeBPEParts(piece []byte, parts []bpePart) []bpePart {
	for len(parts) > 1 {
		bestIndex := -1
		bestRank := int(^uint(0) >> 1)
		for i := 0; i < len(parts)-1; i++ {
			start := parts[i].start
			end := parts[i+1].end
			if rank, ok := e.ranks[string(piece[start:end])]; ok && rank < bestRank {
				bestRank = rank
				bestIndex = i
			}
		}
		if bestIndex < 0 {
			break
		}
		parts[bestIndex].end = parts[bestIndex+1].end
		copy(parts[bestIndex+1:], parts[bestIndex+2:])
		parts = parts[:len(parts)-1]
	}
	return parts
}

func newEngine(name string, data []byte, segmenter Segmenter, specials map[string]int) (*Engine, error) {
	ranks, decoder, err := parseTiktoken(data)
	if err != nil {
		return nil, err
	}
	usedSpecialIDs := make(map[int]string, len(specials))
	for token, id := range specials {
		if previous, exists := usedSpecialIDs[id]; exists {
			return nil, fmt.Errorf("duplicate special token id %d for %q and %q", id, previous, token)
		}
		usedSpecialIDs[id] = token
		if id >= 0 && id < len(decoder) && decoder[id] != nil {
			return nil, fmt.Errorf("special token id %d collides with mergeable rank", id)
		}
		decoder = setDecoderToken(decoder, id, []byte(token))
	}
	return &Engine{name: name, segmenter: segmenter, ranks: ranks, decoder: decoder, specials: specials}, nil
}

func parseTiktoken(data []byte) (map[string]int, [][]byte, error) {
	rows := bytes.Count(data, []byte{'\n'}) + 1
	ranks := make(map[string]int, rows)
	decoder := make([][]byte, rows)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		fields := bytes.Fields(line)
		if len(fields) != 2 {
			return nil, nil, fmt.Errorf("invalid tiktoken row %d", lineNo)
		}
		raw, err := base64.StdEncoding.DecodeString(string(fields[0]))
		if err != nil {
			return nil, nil, fmt.Errorf("decode token row %d: %w", lineNo, err)
		}
		rank, err := strconv.Atoi(string(fields[1]))
		if err != nil {
			return nil, nil, fmt.Errorf("parse rank row %d: %w", lineNo, err)
		}
		if rank < 0 {
			return nil, nil, fmt.Errorf("negative rank row %d", lineNo)
		}
		stable := append([]byte(nil), raw...)
		key := string(stable)
		if _, exists := ranks[key]; exists {
			return nil, nil, fmt.Errorf("duplicate token row %d", lineNo)
		}
		if rank < len(decoder) && decoder[rank] != nil {
			return nil, nil, fmt.Errorf("duplicate rank row %d", lineNo)
		}
		ranks[key] = rank
		decoder = setDecoderToken(decoder, rank, stable)
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	return ranks, decoder, nil
}

func setDecoderToken(decoder [][]byte, id int, raw []byte) [][]byte {
	if id < 0 {
		return decoder
	}
	if id >= len(decoder) {
		grown := make([][]byte, id+1)
		copy(grown, decoder)
		decoder = grown
	}
	decoder[id] = raw
	return decoder
}

func cl100kSpecialTokens() map[string]int {
	return map[string]int{
		"<|endoftext|>":   100257,
		"<|fim_prefix|>":  100258,
		"<|fim_middle|>":  100259,
		"<|fim_suffix|>":  100260,
		"<|endofprompt|>": 100276,
	}
}

func o200kBaseSpecialTokens() map[string]int {
	return map[string]int{
		"<|endoftext|>":   199999,
		"<|endofprompt|>": 200018,
	}
}

func o200kHarmonySpecialTokens() map[string]int {
	specials := map[string]int{
		"<|startoftext|>": 199998,
		"<|endoftext|>":   199999,
		"<|return|>":      200002,
		"<|constrain|>":   200003,
		"<|channel|>":     200005,
		"<|start|>":       200006,
		"<|end|>":         200007,
		"<|message|>":     200008,
		"<|call|>":        200012,
		"<|endofprompt|>": 200018,
	}
	usedIDs := make(map[int]struct{}, len(specials))
	for _, id := range specials {
		usedIDs[id] = struct{}{}
	}
	for i := 200000; i <= 201087; i++ {
		if _, used := usedIDs[i]; used {
			continue
		}
		specials[fmt.Sprintf("<|reserved_%d|>", i)] = i
	}
	return specials
}

// MergeableRanks returns a sorted snapshot of token ranks. It is intended for tests and diagnostics.
func (e *Engine) MergeableRanks() []int {
	if e == nil {
		return nil
	}
	ranks := make([]int, 0, len(e.ranks))
	for _, rank := range e.ranks {
		ranks = append(ranks, rank)
	}
	sort.Ints(ranks)
	return ranks
}
