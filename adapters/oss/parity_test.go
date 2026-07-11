package oss

import (
	"bufio"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type parityCase struct {
	Name   string `json:"name"`
	Text   string `json:"text"`
	Tokens []int  `json:"tokens"`
	Count  *int   `json:"count"`
	Decode string `json:"decode"`
}

func TestParityFixtures(t *testing.T) {
	modelPath := os.Getenv("OMNITOKEN_OSS_SENTENCEPIECE_MODEL")
	fixturesPath := os.Getenv("OMNITOKEN_OSS_PARITY_JSONL")
	if modelPath == "" || fixturesPath == "" {
		t.Skip("set OMNITOKEN_OSS_SENTENCEPIECE_MODEL and OMNITOKEN_OSS_PARITY_JSONL")
	}
	engine, err := NewSentencePiece("oss_parity", Options{Source: ModelSource{Path: modelPath}})
	if err != nil {
		t.Fatal(err)
	}
	runParityFixtures(t, fixturesPath, engine)
}

func runParityFixtures(t *testing.T, path string, engine *Engine) {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		var tc parityCase
		if err := json.Unmarshal(scanner.Bytes(), &tc); err != nil {
			t.Fatalf("%s:%d: %v", path, line, err)
		}
		name := tc.Name
		if name == "" {
			name = tc.Text
		}
		t.Run(name, func(t *testing.T) {
			if tc.Tokens != nil {
				if got := engine.EncodeOrdinary(tc.Text); !reflect.DeepEqual(got, tc.Tokens) {
					t.Fatalf("EncodeOrdinary = %v, want %v", got, tc.Tokens)
				}
			}
			wantCount := len(tc.Tokens)
			if tc.Count != nil {
				wantCount = *tc.Count
			}
			if tc.Tokens != nil || tc.Count != nil {
				if got := engine.CountTokens(tc.Text); got != wantCount {
					t.Fatalf("CountTokens = %d, want %d", got, wantCount)
				}
			}
			if tc.Decode != "" && tc.Tokens != nil {
				if got := engine.Decode(tc.Tokens); got != tc.Decode {
					t.Fatalf("Decode = %q, want %q", got, tc.Decode)
				}
			}
		})
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}
