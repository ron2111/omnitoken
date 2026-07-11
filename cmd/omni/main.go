package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/ron2111/omnitoken"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "count":
		err = runCount(os.Args[2:])
	case "encode":
		err = runEncode(os.Args[2:])
	case "decode":
		err = runDecode(os.Args[2:])
	case "cache":
		err = runCache(os.Args[2:])
	case "bench":
		err = runBench(os.Args[2:])
	case "encodings":
		err = runEncodings(os.Args[2:])
	case "models":
		err = runModels(os.Args[2:])
	case "help", "-h", "--help":
		usage()
		return
	default:
		err = fmt.Errorf("unknown command: %s", os.Args[1])
	}
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	_, _ = fmt.Fprint(os.Stderr, `OmniToken CLI

Usage:
  omni count  [-model gpt-4o|-encoding o200k_base] [-file path] [text]
  omni encode [-model gpt-4o|-encoding o200k_base] [-file path] [text]
  omni decode [-model gpt-4o|-encoding o200k_base] [token ids]
  omni cache  [-model gpt-4o|-encoding o200k_base] [-profile openai|generic] [-file path] [text]
  omni bench  -input path -timings dir -name name [-model cl100k_base] [-iters 100] [-warmup 10]
  omni encodings [-json]
  omni models [-json] [-prefixes]

Examples:
  omni count -model gpt-4o "hello world"
  omni encode -encoding o200k_base "hello world"
  omni decode -encoding o200k_base "24912 2375"
  omni cache -model gpt-4o -profile openai system-prompt.txt
  omni encodings
  omni models -prefixes
`)
}

type engineFlags struct {
	model    string
	encoding string
	file     string
}

func addEngineFlags(fs *flag.FlagSet, flags *engineFlags) {
	fs.StringVar(&flags.model, "model", "gpt-4o", "model name")
	fs.StringVar(&flags.encoding, "encoding", "", "encoding name; overrides model")
	fs.StringVar(&flags.file, "file", "", "read input text from file")
}

func engineFromFlags(flags engineFlags) (omnitoken.ModelEngine, error) {
	if flags.encoding != "" {
		return omnitoken.ForEncoding(resolveEncoding(flags.encoding))
	}
	return omnitoken.ForModel(flags.model)
}

func runCount(args []string) error {
	fs := flag.NewFlagSet("count", flag.ExitOnError)
	var flags engineFlags
	addEngineFlags(fs, &flags)
	if err := fs.Parse(args); err != nil {
		return err
	}
	engine, err := engineFromFlags(flags)
	if err != nil {
		return err
	}
	text, err := inputText(flags.file, fs.Args())
	if err != nil {
		return err
	}
	fmt.Println(engine.CountTokens(text))
	return nil
}

func runEncode(args []string) error {
	fs := flag.NewFlagSet("encode", flag.ExitOnError)
	var flags engineFlags
	addEngineFlags(fs, &flags)
	if err := fs.Parse(args); err != nil {
		return err
	}
	engine, err := engineFromFlags(flags)
	if err != nil {
		return err
	}
	text, err := inputText(flags.file, fs.Args())
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(engine.EncodeOrdinary(text))
	if err != nil {
		return err
	}
	fmt.Println(string(encoded))
	return nil
}

func runDecode(args []string) error {
	fs := flag.NewFlagSet("decode", flag.ExitOnError)
	var flags engineFlags
	addEngineFlags(fs, &flags)
	if err := fs.Parse(args); err != nil {
		return err
	}
	engine, err := engineFromFlags(flags)
	if err != nil {
		return err
	}
	raw, err := inputText(flags.file, fs.Args())
	if err != nil {
		return err
	}
	tokens, err := parseTokenIDs(raw)
	if err != nil {
		return err
	}
	fmt.Println(engine.Decode(tokens))
	return nil
}

func runCache(args []string) error {
	fs := flag.NewFlagSet("cache", flag.ExitOnError)
	var flags engineFlags
	profileName := "openai"
	blockSize := 0
	minimumTokens := -1
	addEngineFlags(fs, &flags)
	fs.StringVar(&profileName, "profile", profileName, "cache profile: openai, generic, custom")
	fs.IntVar(&blockSize, "block", blockSize, "custom block size")
	fs.IntVar(&minimumTokens, "min", minimumTokens, "custom minimum tokens")
	if err := fs.Parse(args); err != nil {
		return err
	}
	engine, err := engineFromFlags(flags)
	if err != nil {
		return err
	}
	text, err := inputText(flags.file, fs.Args())
	if err != nil {
		return err
	}
	profile := cacheProfile(profileName)
	if blockSize > 0 {
		profile.BlockSize = blockSize
	}
	if minimumTokens >= 0 {
		profile.MinimumTokens = minimumTokens
	}
	report := omnitoken.NewCacheAligner(engine).AlignPromptToProfile(text, profile)
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(encoded))
	return nil
}

func runBench(args []string) error {
	fs := flag.NewFlagSet("bench", flag.ExitOnError)
	var name string
	var model string
	var input string
	var timings string
	var iterations int
	var warmup int
	var batch int
	fs.StringVar(&name, "name", "omnitoken", "benchmark timing name")
	fs.StringVar(&model, "model", "cl100k_base", "model/encoding name or tokenizer file path")
	fs.StringVar(&input, "input", "", "UTF-8 input file")
	fs.StringVar(&timings, "timings", "timings", "output timing directory")
	fs.IntVar(&iterations, "iters", 100, "measured iterations")
	fs.IntVar(&warmup, "warmup", 10, "warmup iterations")
	fs.IntVar(&batch, "batch", 10, "encodes per timing iteration")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if input == "" {
		return fmt.Errorf("input file is required")
	}
	if iterations < 0 || warmup < 0 || batch <= 0 {
		return fmt.Errorf("invalid iteration settings")
	}

	textBytes, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	engine, err := omnitoken.ForEncoding(resolveEncoding(model))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(timings, 0o755); err != nil {
		return err
	}
	out, err := os.Create(filepath.Join(timings, name+".txt"))
	if err != nil {
		return err
	}
	defer out.Close()

	debug.SetGCPercent(-1)
	runtime.GC()

	text := string(textBytes)
	var sink []int
	for i := 0; i < iterations+warmup; i++ {
		start := time.Now()
		for j := 0; j < batch; j++ {
			sink = engine.EncodeOrdinary(text)
		}
		elapsed := time.Since(start)
		if i >= warmup {
			_, _ = fmt.Fprintf(out, "%.12f\n", elapsed.Seconds())
		}
	}
	if len(sink) == 0 && text != "" {
		return fmt.Errorf("empty token output for non-empty input")
	}
	return nil
}

func runEncodings(args []string) error {
	fs := flag.NewFlagSet("encodings", flag.ExitOnError)
	jsonOutput := false
	fs.BoolVar(&jsonOutput, "json", jsonOutput, "emit JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	encodings := omnitoken.RegisteredEncodings()
	if jsonOutput {
		encoded, err := json.MarshalIndent(encodings, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
		return nil
	}
	for _, encoding := range encodings {
		fmt.Println(encoding)
	}
	return nil
}

func runModels(args []string) error {
	fs := flag.NewFlagSet("models", flag.ExitOnError)
	jsonOutput := false
	includePrefixes := false
	fs.BoolVar(&jsonOutput, "json", jsonOutput, "emit JSON")
	fs.BoolVar(&includePrefixes, "prefixes", includePrefixes, "include prefix mappings")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rows := modelRows(includePrefixes)
	if jsonOutput {
		encoded, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
		return nil
	}
	for _, row := range rows {
		fmt.Printf("%s\t%s\t%s\t%s\n", row.Type, row.Provider, row.Name, row.Encoding)
	}
	return nil
}

type modelRow struct {
	Type     string             `json:"type"`
	Provider omnitoken.Provider `json:"provider"`
	Name     string             `json:"name"`
	Encoding string             `json:"encoding"`
}

func modelRows(includePrefixes bool) []modelRow {
	models := omnitoken.RegisteredModels()
	rows := make([]modelRow, 0, len(models)+len(omnitoken.RegisteredModelPrefixes()))
	for _, model := range models {
		rows = append(rows, modelRow{Type: "exact", Provider: model.Provider, Name: model.Model, Encoding: model.Encoding})
	}
	if includePrefixes {
		for _, prefix := range omnitoken.RegisteredModelPrefixes() {
			rows = append(rows, modelRow{Type: "prefix", Provider: prefix.Provider, Name: prefix.Prefix, Encoding: prefix.Encoding})
		}
	}
	return rows
}

func inputText(path string, args []string) (string, error) {
	if path != "" {
		data, err := os.ReadFile(path)
		return string(data), err
	}
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	data, err := io.ReadAll(os.Stdin)
	return string(data), err
}

func parseTokenIDs(text string) ([]int, error) {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "[")
	text = strings.TrimSuffix(text, "]")
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\n' || r == '\r' || r == '\t'
	})
	tokens := make([]int, 0, len(fields))
	for _, field := range fields {
		if field == "" {
			continue
		}
		id, err := strconv.Atoi(field)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, id)
	}
	return tokens, nil
}

func resolveEncoding(model string) string {
	base := filepath.Base(model)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	switch base {
	case "cl100k", "cl100k_base":
		return omnitoken.EncodingCL100KBase
	case "o200k", "o200k_base":
		return omnitoken.EncodingO200KBase
	case "o200k_harmony":
		return omnitoken.EncodingO200KHarmony
	default:
		return model
	}
}

func cacheProfile(name string) omnitoken.CacheProfile {
	switch name {
	case "generic":
		return omnitoken.CacheProfileGeneric
	case "custom":
		return omnitoken.CacheProfile{Name: "custom", BlockSize: 1024}
	default:
		return omnitoken.CacheProfileOpenAI
	}
}
