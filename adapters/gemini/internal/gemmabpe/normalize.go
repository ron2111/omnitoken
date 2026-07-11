package gemmabpe

import (
	"strings"
	"unsafe"
)

// normalize performs unicode normalization.
//
// SentencePiece has a feature to perform configurable unicode normalization on
// the input text and has some options for adding dummy whitespace prefixes or
// trimming whitespace. However, the model we're working with has a very simple
// normalizer that does none of this. These options can be added in the future
// if needed.
func normalize(text string) string {
	return replaceSpacesBySeparator(text)
}

const whitespaceSeparator = "▁"

// replaceSpacesBySeparator replaces spaces by the whitespace separator used by
// the model.
func replaceSpacesBySeparator(text string) string {
	return strings.ReplaceAll(text, " ", whitespaceSeparator)
}

func normalizeInto(text string, scratch []byte) (string, []byte) {
	if !strings.Contains(text, " ") {
		return text, scratch[:0]
	}
	spaces := strings.Count(text, " ")
	need := len(text) + spaces*(len(whitespaceSeparator)-1)
	if cap(scratch) < need {
		scratch = make([]byte, 0, need)
	}
	scratch = scratch[:0]
	for {
		idx := strings.IndexByte(text, ' ')
		if idx < 0 {
			scratch = append(scratch, text...)
			break
		}
		scratch = append(scratch, text[:idx]...)
		scratch = append(scratch, whitespaceSeparator...)
		text = text[idx+1:]
	}
	if len(scratch) == 0 {
		return "", scratch
	}
	return unsafe.String(unsafe.SliceData(scratch), len(scratch)), scratch
}

// replaceSeparatorsBySpace replaces the whitespace separator used by
// the model back with spaces.
func replaceSeparatorsBySpace(text string) string {
	return strings.ReplaceAll(text, whitespaceSeparator, " ")
}
