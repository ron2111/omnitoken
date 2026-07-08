package omnitoken

import (
	"bytes"
	"encoding/base64"
	"fmt"
)

func parseBPERanks(data []byte) (map[string]uint32, [][]byte, error) {
	rows := bytes.Count(data, []byte{'\n'}) + 1
	ranks := make(map[string]uint32, rows)
	decoder := make([][]byte, rows)
	for lineNo, start := 1, 0; start < len(data); lineNo++ {
		end := bytes.IndexByte(data[start:], '\n')
		var line []byte
		if end < 0 {
			line = data[start:]
			start = len(data)
		} else {
			line = data[start : start+end]
			start += end + 1
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		sep := bytes.IndexByte(line, ' ')
		if sep < 0 {
			return nil, nil, fmt.Errorf("invalid BPE rank row %d", lineNo)
		}
		encoded := line[:sep]
		rankBytes := bytes.TrimSpace(line[sep+1:])
		if len(encoded) == 0 || len(rankBytes) == 0 || bytes.IndexByte(rankBytes, ' ') >= 0 {
			return nil, nil, fmt.Errorf("invalid BPE rank row %d", lineNo)
		}
		stable := make([]byte, base64.StdEncoding.DecodedLen(len(encoded)))
		n, err := base64.StdEncoding.Decode(stable, encoded)
		if err != nil {
			return nil, nil, fmt.Errorf("decode token row %d: %w", lineNo, err)
		}
		stable = stable[:n]
		rank, err := parseRank(rankBytes)
		if err != nil {
			return nil, nil, fmt.Errorf("parse rank row %d: %w", lineNo, err)
		}
		// The rank key and decoder entry share the same immutable token bytes.
		key := unsafeBytesString(stable)
		if _, exists := ranks[key]; exists {
			return nil, nil, fmt.Errorf("duplicate token row %d", lineNo)
		}
		if rank < len(decoder) && decoder[rank] != nil {
			return nil, nil, fmt.Errorf("duplicate rank row %d", lineNo)
		}
		if rank > int(^uint32(0)) {
			return nil, nil, fmt.Errorf("rank row %d overflows uint32", lineNo)
		}
		ranks[key] = uint32(rank)
		decoder = setDecoderToken(decoder, rank, stable)
	}
	return ranks, decoder, nil
}

func parseRank(src []byte) (int, error) {
	if len(src) == 0 {
		return 0, fmt.Errorf("empty rank")
	}
	maxInt := int(^uint(0) >> 1)
	n := 0
	for _, c := range src {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid rank")
		}
		digit := int(c - '0')
		if n > (maxInt-digit)/10 {
			return 0, fmt.Errorf("rank overflows int")
		}
		n = n*10 + digit
	}
	return n, nil
}

func setDecoderToken(decoder [][]byte, id int, raw []byte) [][]byte {
	if id < 0 {
		return decoder
	}
	decoder = growDecoder(decoder, id)
	decoder[id] = raw
	return decoder
}

func growDecoder(decoder [][]byte, id int) [][]byte {
	if id < len(decoder) {
		return decoder
	}
	grown := make([][]byte, id+1)
	copy(grown, decoder)
	return grown
}
