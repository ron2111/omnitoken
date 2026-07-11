package cacheflow

import "strings"

// DetectBreakers returns best-effort warnings for dynamic data commonly placed
// before stable prompt content. These are heuristics, not provider guarantees.
func DetectBreakers(items []TraceItem) []CacheBreaker {
	var out []CacheBreaker
	for _, item := range items {
		if len(item.Parts) > 0 {
			for _, part := range item.Parts {
				if part.Stable {
					continue
				}
				out = append(out, detectTextBreakers(item.ID, part.Name, part.Text)...)
			}
			continue
		}
		out = append(out, detectTextBreakers(item.ID, "", item.Prompt)...)
	}
	return out
}

func detectTextBreakers(itemID string, part string, text string) []CacheBreaker {
	checks := []struct {
		kind string
		ok   bool
		msg  string
	}{
		{"timestamp", hasLikelyTimestamp(text), "Likely timestamp detected; keep dynamic timestamps after stable cacheable prompt sections."},
		{"uuid", hasLikelyUUID(text), "Likely UUID/request ID detected; keep per-request identifiers after stable cacheable prompt sections."},
		{"random_json", hasLikelyDynamicJSON(text), "Likely dynamic JSON metadata detected; keep changing metadata after stable instructions and context."},
	}
	out := make([]CacheBreaker, 0, len(checks))
	for _, check := range checks {
		if check.ok {
			out = append(out, CacheBreaker{ItemID: itemID, Part: part, Kind: check.kind, Message: check.msg})
		}
	}
	return out
}

func hasLikelyTimestamp(text string) bool {
	for i := 0; i+19 <= len(text); i++ {
		if isDigit(text[i]) && isDigit(text[i+1]) && isDigit(text[i+2]) && isDigit(text[i+3]) &&
			text[i+4] == '-' && isDigit(text[i+5]) && isDigit(text[i+6]) && text[i+7] == '-' &&
			isDigit(text[i+8]) && isDigit(text[i+9]) && (text[i+10] == 'T' || text[i+10] == ' ') &&
			isDigit(text[i+11]) && isDigit(text[i+12]) && text[i+13] == ':' &&
			isDigit(text[i+14]) && isDigit(text[i+15]) && text[i+16] == ':' &&
			isDigit(text[i+17]) && isDigit(text[i+18]) {
			return true
		}
	}
	return false
}

func hasLikelyUUID(text string) bool {
	for i := 0; i+36 <= len(text); i++ {
		if isHexRun(text[i:i+8]) && text[i+8] == '-' && isHexRun(text[i+9:i+13]) && text[i+13] == '-' &&
			isHexRun(text[i+14:i+18]) && text[i+18] == '-' && isHexRun(text[i+19:i+23]) && text[i+23] == '-' &&
			isHexRun(text[i+24:i+36]) {
			return true
		}
	}
	return false
}

func hasLikelyDynamicJSON(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "\"request_id\"") || strings.Contains(lower, "\"timestamp\"") || strings.Contains(lower, "\"trace_id\"") || strings.Contains(lower, "\"session_id\"")
}

func isDigit(b byte) bool { return '0' <= b && b <= '9' }

func isHexRun(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		b := s[i]
		if !isDigit(b) && !(('a' <= b && b <= 'f') || ('A' <= b && b <= 'F')) {
			return false
		}
	}
	return true
}
