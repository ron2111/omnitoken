package omnitoken

import "strconv"

func cloneSpecials(specials map[string]int) map[string]int {
	if len(specials) == 0 {
		return nil
	}
	out := make(map[string]int, len(specials))
	for token, id := range specials {
		out[token] = id
	}
	return out
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
	for i := 200000; i <= 200001; i++ {
		specials[reservedTokenName(i)] = i
	}
	specials["<|reserved_200004|>"] = 200004
	for i := 200009; i <= 200011; i++ {
		specials[reservedTokenName(i)] = i
	}
	for i := 200013; i <= 201087; i++ {
		specials[reservedTokenName(i)] = i
	}
	return specials
}

func reservedTokenName(id int) string {
	return "<|reserved_" + strconv.Itoa(id) + "|>"
}
