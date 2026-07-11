package omnitoken

import "fmt"

const (
	SegmenterCL100K = "cl100k"
	SegmenterO200K  = "o200k"
)

// ByteBPEOptions configures a byte-level BPE engine from tiktoken-style ranks.
type ByteBPEOptions struct {
	Name            string
	Data            []byte
	Segmenter       string
	CustomSegmenter Segmenter
	Specials        map[string]int
}

// NewByteBPE builds a pure-Go byte-level BPE engine from tiktoken-style data.
func NewByteBPE(opts ByteBPEOptions) (*Engine, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("omnitoken: BPE name is required")
	}
	if len(opts.Data) == 0 {
		return nil, fmt.Errorf("omnitoken: BPE data is required")
	}
	segmenter := opts.CustomSegmenter
	if segmenter == nil {
		var err error
		segmenter, err = segmenterByName(opts.Segmenter)
		if err != nil {
			return nil, err
		}
	}
	return newEngine(opts.Name, opts.Data, segmenter, opts.Specials)
}

func segmenterByName(name string) (Segmenter, error) {
	switch name {
	case "", SegmenterCL100K:
		return segmenterFunc(nextCL100K), nil
	case SegmenterO200K:
		return segmenterFunc(nextO200K), nil
	default:
		return nil, fmt.Errorf("omnitoken: unsupported BPE segmenter: %s", name)
	}
}
