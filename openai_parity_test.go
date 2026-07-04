package omnitoken

import (
	"reflect"
	"testing"
)

func TestOpenAIParitySmoke(t *testing.T) {
	tests := []struct {
		encoding string
		text     string
		want     []int
	}{
		{EncodingCL100KBase, "hello world", []int{15339, 1917}},
		{EncodingCL100KBase, "Hello, world!", []int{9906, 11, 1917, 0}},
		{EncodingCL100KBase, "I'm testing 123456 tokens.", []int{40, 2846, 7649, 220, 4513, 10961, 11460, 13}},
		{EncodingCL100KBase, "  hello", []int{220, 24748}},
		{EncodingCL100KBase, "hello  world", []int{15339, 220, 1917}},
		{EncodingCL100KBase, "こんにちは世界", []int{90115, 3574, 244, 98220}},
		{EncodingCL100KBase, "😀 test", []int{76460, 222, 1296}},
		{EncodingCL100KBase, "foo_bar/baz\n", []int{8134, 14725, 3554, 1394, 198}},
		{EncodingCL100KBase, "HTTPServerError", []int{9412, 39609}},
		{EncodingCL100KBase, "ABCdefGHI", []int{26484, 755, 38, 24860}},
		{EncodingCL100KBase, "<|endoftext|>", []int{27, 91, 8862, 728, 428, 91, 29}},

		{EncodingO200KBase, "hello world", []int{24912, 2375}},
		{EncodingO200KBase, "Hello, world!", []int{13225, 11, 2375, 0}},
		{EncodingO200KBase, "I'm testing 123456 tokens.", []int{15390, 11493, 220, 7633, 19354, 20290, 13}},
		{EncodingO200KBase, "  hello", []int{220, 40617}},
		{EncodingO200KBase, "hello  world", []int{24912, 220, 2375}},
		{EncodingO200KBase, "こんにちは世界", []int{95839, 28428}},
		{EncodingO200KBase, "😀 test", []int{84083, 1746}},
		{EncodingO200KBase, "foo_bar/baz\n", []int{16660, 31828, 7611, 1071, 198}},
		{EncodingO200KBase, "HTTPServerError", []int{17893, 6444, 2255}},
		{EncodingO200KBase, "ABCdefGHI", []int{44197, 1314, 38, 36525}},
		{EncodingO200KBase, "<|endoftext|>", []int{27, 91, 419, 1440, 919, 91, 29}},

		{EncodingO200KHarmony, "hello world", []int{24912, 2375}},
		{EncodingO200KHarmony, "Hello, world!", []int{13225, 11, 2375, 0}},
		{EncodingO200KHarmony, "I'm testing 123456 tokens.", []int{15390, 11493, 220, 7633, 19354, 20290, 13}},
		{EncodingO200KHarmony, "<|endoftext|>", []int{27, 91, 419, 1440, 919, 91, 29}},
	}

	for _, tt := range tests {
		t.Run(tt.encoding+"/"+tt.text, func(t *testing.T) {
			engine, err := ForEncoding(tt.encoding)
			if err != nil {
				t.Fatal(err)
			}
			got := engine.EncodeOrdinary(tt.text)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("EncodeOrdinary(%q) = %v, want %v", tt.text, got, tt.want)
			}
			if count := engine.CountTokens(tt.text); count != len(tt.want) {
				t.Fatalf("CountTokens(%q) = %d, want %d", tt.text, count, len(tt.want))
			}
		})
	}
}

func TestOpenAIParityEdgeCases(t *testing.T) {
	tests := []struct {
		encoding string
		text     string
		want     []int
	}{
		{EncodingCL100KBase, " ", []int{220}},
		{EncodingCL100KBase, "  ", []int{256}},
		{EncodingCL100KBase, "   hello", []int{256, 24748}},
		{EncodingCL100KBase, "\n", []int{198}},
		{EncodingCL100KBase, " \n", []int{720}},
		{EncodingCL100KBase, " \n\nx", []int{4815, 87}},
		{EncodingCL100KBase, "1234", []int{4513, 19}},
		{EncodingCL100KBase, "1234567", []int{4513, 10961, 22}},
		{EncodingCL100KBase, "don't you're I'M", []int{15357, 956, 499, 2351, 358, 28703}},
		{EncodingCL100KBase, "snake_case", []int{73239, 19640}},
		{EncodingCL100KBase, "[]{}() /path/to/file.go", []int{1318, 6390, 368, 611, 2398, 33529, 24849, 18487}},
		{EncodingCL100KBase, "a\U0001f680b", []int{64, 9468, 248, 222, 65}},
		{EncodingCL100KBase, "\u4e2d\u6587test", []int{16325, 17161, 1985}},

		{EncodingO200KBase, " ", []int{220}},
		{EncodingO200KBase, "  ", []int{256}},
		{EncodingO200KBase, "   hello", []int{256, 40617}},
		{EncodingO200KBase, "\n", []int{198}},
		{EncodingO200KBase, " \n", []int{793}},
		{EncodingO200KBase, " \n\nx", []int{1202, 87}},
		{EncodingO200KBase, "1234", []int{7633, 19}},
		{EncodingO200KBase, "1234567", []int{7633, 19354, 22}},
		{EncodingO200KBase, "don't you're I'M", []int{91418, 7163, 3413, 44}},
		{EncodingO200KBase, "snake_case", []int{162012, 43667}},
		{EncodingO200KBase, "[]{}() /path/to/file.go", []int{1951, 12083, 416, 820, 4189, 72231, 51766, 32812}},
		{EncodingO200KBase, "a\U0001f680b", []int{64, 112927, 222, 65}},
		{EncodingO200KBase, "\u4e2d\u6587test", []int{10667, 3190}},
	}
	for _, tt := range tests {
		t.Run(tt.encoding+"/edge", func(t *testing.T) {
			engine, err := ForEncoding(tt.encoding)
			if err != nil {
				t.Fatal(err)
			}
			got := engine.EncodeOrdinary(tt.text)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("EncodeOrdinary(%q) = %v, want %v", tt.text, got, tt.want)
			}
			if count := engine.CountTokens(tt.text); count != len(tt.want) {
				t.Fatalf("CountTokens(%q) = %d, want %d", tt.text, count, len(tt.want))
			}
		})
	}
}
func TestRegistryMappings(t *testing.T) {
	tests := map[string]string{
		"gpt-4o":                 EncodingO200KBase,
		"gpt-4o-2024-08-06":      EncodingO200KBase,
		"gpt-4.1":                EncodingO200KBase,
		"gpt-5-2026-01-01":       EncodingO200KBase,
		"gpt-4":                  EncodingCL100KBase,
		"gpt-3.5-turbo-0125":     EncodingCL100KBase,
		"text-embedding-3-large": EncodingCL100KBase,
		"gpt-oss-120b":           EncodingO200KHarmony,
	}

	for model, want := range tests {
		engine, err := ForModel(model)
		if err != nil {
			t.Fatalf("ForModel(%q): %v", model, err)
		}
		got := engine.(*Engine).Encoding()
		if got != want {
			t.Fatalf("ForModel(%q) encoding = %q, want %q", model, got, want)
		}
	}
}

func TestDecodeRoundTrip(t *testing.T) {
	texts := []string{
		"hello world",
		"Hello, world!",
		"こんにちは世界",
		"😀 test",
		"foo_bar/baz\n",
	}
	for _, encoding := range []string{EncodingCL100KBase, EncodingO200KBase, EncodingO200KHarmony} {
		engine, err := ForEncoding(encoding)
		if err != nil {
			t.Fatal(err)
		}
		for _, text := range texts {
			if got := engine.Decode(engine.EncodeOrdinary(text)); got != text {
				t.Fatalf("%s round-trip = %q, want %q", encoding, got, text)
			}
		}
	}
}

func TestSpecialDecode(t *testing.T) {
	cl, err := ForEncoding(EncodingCL100KBase)
	if err != nil {
		t.Fatal(err)
	}
	if got := cl.Decode([]int{100257}); got != "<|endoftext|>" {
		t.Fatalf("cl100k special decode = %q", got)
	}

	harmony, err := ForEncoding(EncodingO200KHarmony)
	if err != nil {
		t.Fatal(err)
	}
	if got := harmony.Decode([]int{200006, 200005, 200007}); got != "<|start|><|channel|><|end|>" {
		t.Fatalf("harmony special decode = %q", got)
	}
}
