package omnitoken

import (
	"unicode"
	"unicode/utf8"
)

func nextCL100K(src []byte, start int) int {
	if end, ok := contractionEnd(src, start, false); ok {
		return end
	}

	r, size := decodeRune(src, start)
	if isOptionalWordPrefix(r) {
		next := start + size
		if next < len(src) {
			nr, _ := decodeRune(src, next)
			if isLetter(nr) {
				return consumeLetters(src, next)
			}
		}
	}
	if isLetter(r) {
		return consumeLetters(src, start)
	}
	if isNumber(r) {
		return consumeNumbers(src, start, 3)
	}

	if r == ' ' {
		next := start + size
		if next < len(src) {
			nr, _ := decodeRune(src, next)
			if isPunctuationForToken(nr) {
				return consumePunctuationAndNewlines(src, next)
			}
		}
	}
	if isPunctuationForToken(r) {
		return consumePunctuationAndNewlines(src, start)
	}

	if isWhitespace(r) {
		return consumeCL100KWhitespace(src, start)
	}
	return start + size
}

func nextO200K(src []byte, start int) int {
	if end, ok := contractionEnd(src, start, true); ok {
		return end
	}

	r, size := decodeRune(src, start)
	wordStart := start
	if isOptionalWordPrefix(r) {
		next := start + size
		if next < len(src) {
			nr, _ := decodeRune(src, next)
			if isO200KWordChar(nr) {
				wordStart = next
			}
		}
	}
	if wordStart != start || isO200KWordChar(r) {
		if end, ok := consumeO200KWord(src, wordStart); ok {
			return consumeOptionalContraction(src, end, true)
		}
	}
	if isNumber(r) {
		return consumeNumbers(src, start, 3)
	}

	if r == ' ' {
		next := start + size
		if next < len(src) {
			nr, _ := decodeRune(src, next)
			if isPunctuationForToken(nr) {
				return consumeO200KPunctuationTail(src, consumePunctuationRun(src, next))
			}
		}
	}
	if isPunctuationForToken(r) {
		return consumeO200KPunctuationTail(src, consumePunctuationRun(src, start))
	}

	if isWhitespace(r) {
		return consumeO200KWhitespace(src, start)
	}
	return start + size
}

func contractionEnd(src []byte, start int, includeD bool) (int, bool) {
	if start >= len(src) || src[start] != '\'' {
		return 0, false
	}
	suffixes := contractionSuffixesNoD[:]
	if includeD {
		suffixes = contractionSuffixesWithD[:]
	}
	for _, suffix := range suffixes {
		if hasFoldedASCIIPrefix(src[start+1:], suffix) {
			return start + 1 + len(suffix), true
		}
	}
	return 0, false
}

func consumeOptionalContraction(src []byte, start int, includeD bool) int {
	if end, ok := contractionEnd(src, start, includeD); ok {
		return end
	}
	return start
}

var contractionSuffixesWithD = [...]string{"ll", "ve", "re", "s", "t", "m", "d"}
var contractionSuffixesNoD = [...]string{"ll", "ve", "re", "s", "d", "m", "t"}

func hasFoldedASCIIPrefix(src []byte, suffix string) bool {
	if len(src) < len(suffix) {
		return false
	}
	for i := 0; i < len(suffix); i++ {
		b := src[i]
		if 'A' <= b && b <= 'Z' {
			b += 'a' - 'A'
		}
		if b != suffix[i] {
			return false
		}
	}
	return true
}

func consumeLetters(src []byte, start int) int {
	for i := start; i < len(src); {
		if src[i] < utf8.RuneSelf {
			if !isASCIILetter(src[i]) {
				return i
			}
			i++
			continue
		}
		r, size := decodeRune(src, i)
		if !isLetter(r) {
			return i
		}
		i += size
	}
	return len(src)
}

func consumeNumbers(src []byte, start int, maxRunes int) int {
	count := 0
	for i := start; i < len(src) && count < maxRunes; {
		if src[i] < utf8.RuneSelf {
			if !isASCIIDigit(src[i]) {
				return i
			}
			i++
			count++
			if count == maxRunes {
				return i
			}
			continue
		}
		r, size := decodeRune(src, i)
		if !isNumber(r) {
			return i
		}
		i += size
		count++
		if count == maxRunes {
			return i
		}
	}
	if count == 0 {
		_, size := decodeRune(src, start)
		return start + size
	}
	return len(src)
}

func consumePunctuationRun(src []byte, start int) int {
	for i := start; i < len(src); {
		if src[i] < utf8.RuneSelf {
			if !isASCIIPunctuationForToken(src[i]) {
				return i
			}
			i++
			continue
		}
		r, size := decodeRune(src, i)
		if !isPunctuationForToken(r) {
			return i
		}
		i += size
	}
	return len(src)
}

func consumePunctuationAndNewlines(src []byte, start int) int {
	i := consumePunctuationRun(src, start)
	for i < len(src) {
		r, size := decodeRune(src, i)
		if r != '\r' && r != '\n' {
			break
		}
		i += size
	}
	return i
}

func consumeO200KPunctuationTail(src []byte, start int) int {
	for i := start; i < len(src); {
		r, size := decodeRune(src, i)
		if r != '\r' && r != '\n' && r != '/' {
			return i
		}
		i += size
	}
	return len(src)
}

func consumeCL100KWhitespace(src []byte, start int) int {
	runEnd, lastStart, hasNewline, lastNewlineEnd := whitespaceRunInfo(src, start)
	if runEnd == len(src) {
		return runEnd
	}
	if hasNewline {
		return lastNewlineEnd
	}
	if lastStart > start {
		return lastStart
	}
	_, size := decodeRune(src, start)
	return start + size
}

func consumeO200KWhitespace(src []byte, start int) int {
	runEnd, lastStart, hasNewline, lastNewlineEnd := whitespaceRunInfo(src, start)
	if hasNewline {
		return lastNewlineEnd
	}
	if runEnd == len(src) {
		return runEnd
	}
	if lastStart > start {
		return lastStart
	}
	return runEnd
}

func whitespaceRunInfo(src []byte, start int) (runEnd int, lastStart int, hasNewline bool, lastNewlineEnd int) {
	lastStart = start
	for i := start; i < len(src); {
		if src[i] < utf8.RuneSelf {
			if !isASCIIWhitespace(src[i]) {
				return i, lastStart, hasNewline, lastNewlineEnd
			}
			lastStart = i
			if src[i] == '\r' || src[i] == '\n' {
				hasNewline = true
				lastNewlineEnd = i + 1
			}
			i++
			continue
		}
		r, size := decodeRune(src, i)
		if !isWhitespace(r) {
			return i, lastStart, hasNewline, lastNewlineEnd
		}
		lastStart = i
		i += size
		if r == '\r' || r == '\n' {
			hasNewline = true
			lastNewlineEnd = i
		}
	}
	return len(src), lastStart, hasNewline, lastNewlineEnd
}

func consumeO200KWord(src []byte, start int) (int, bool) {
	if start >= len(src) {
		return start, false
	}

	i := start
	for i < len(src) {
		if src[i] < utf8.RuneSelf {
			if !isASCIIO200KUpperGroup(src[i]) || isASCIIO200KLowerGroup(src[i]) {
				break
			}
			i++
			continue
		}
		r, size := decodeRune(src, i)
		if !isO200KUpperGroup(r) {
			break
		}
		if isO200KLowerGroup(r) {
			break
		}
		i += size
	}

	lowerStart := i
	for i < len(src) {
		if src[i] < utf8.RuneSelf {
			if !isASCIIO200KLowerGroup(src[i]) {
				break
			}
			i++
			continue
		}
		r, size := decodeRune(src, i)
		if !isO200KLowerGroup(r) {
			break
		}
		i += size
	}
	if i > lowerStart {
		return i, true
	}

	i = start
	for i < len(src) {
		if src[i] < utf8.RuneSelf {
			if !isASCIIO200KUpperGroup(src[i]) {
				break
			}
			i++
			continue
		}
		r, size := decodeRune(src, i)
		if !isO200KUpperGroup(r) {
			break
		}
		i += size
	}
	if i == start {
		return start, false
	}
	for i < len(src) {
		if src[i] < utf8.RuneSelf {
			if !isASCIIO200KLowerGroup(src[i]) {
				break
			}
			i++
			continue
		}
		r, size := decodeRune(src, i)
		if !isO200KLowerGroup(r) {
			break
		}
		i += size
	}
	return i, true
}

func nextRuneIndex(src []byte, start int) int {
	_, size := decodeRune(src, start)
	return start + size
}

func decodeRune(src []byte, start int) (rune, int) {
	if start >= len(src) {
		return utf8.RuneError, 0
	}
	if src[start] < utf8.RuneSelf {
		return rune(src[start]), 1
	}
	r, size := utf8.DecodeRune(src[start:])
	if r == utf8.RuneError && size == 0 {
		return rune(src[start]), 1
	}
	return r, size
}

func isOptionalWordPrefix(r rune) bool {
	return r != '\r' && r != '\n' && !isLetter(r) && !isNumber(r)
}

func isPunctuationForToken(r rune) bool {
	return !isWhitespace(r) && !isLetter(r) && !isNumber(r)
}

func isLetter(r rune) bool { return unicode.IsLetter(r) }

func isNumber(r rune) bool { return unicode.IsNumber(r) }

func isWhitespace(r rune) bool { return unicode.IsSpace(r) }

func isASCIILetter(b byte) bool { return ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') }

func isASCIIDigit(b byte) bool { return '0' <= b && b <= '9' }

func isASCIIWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\v' || b == '\f' || b == '\r'
}

func isASCIIPunctuationForToken(b byte) bool {
	return !isASCIIWhitespace(b) && !isASCIILetter(b) && !isASCIIDigit(b)
}

func isASCIIO200KUpperGroup(b byte) bool { return 'A' <= b && b <= 'Z' }

func isASCIIO200KLowerGroup(b byte) bool { return 'a' <= b && b <= 'z' }

func isO200KWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsMark(r)
}

func isO200KUpperGroup(r rune) bool {
	return unicode.Is(unicode.Lu, r) || unicode.Is(unicode.Lt, r) || unicode.Is(unicode.Lm, r) || unicode.Is(unicode.Lo, r) || unicode.IsMark(r)
}

func isO200KLowerGroup(r rune) bool {
	return unicode.Is(unicode.Ll, r) || unicode.Is(unicode.Lm, r) || unicode.Is(unicode.Lo, r) || unicode.IsMark(r)
}
