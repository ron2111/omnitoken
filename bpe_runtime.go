package omnitoken

func (e *Engine) appendPieceTokens(dst []int, piece []byte) []int {
	if rank, ok := e.ranks[string(piece)]; ok {
		return append(dst, int(rank))
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
			dst = append(dst, int(rank))
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
		bestRank := uint32(^uint32(0))
		for i := range len(parts) - 1 {
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
