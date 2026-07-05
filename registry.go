package omnitoken

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

const (
	EncodingCL100KBase   = "cl100k_base"
	EncodingO200KBase    = "o200k_base"
	EncodingO200KHarmony = "o200k_harmony"
)

// Provider identifies the ecosystem a model mapping belongs to.
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderGoogle    Provider = "google"
	ProviderMeta      Provider = "meta"
	ProviderCustom    Provider = "custom"
)

// ModelInfo describes how a model name resolves inside the registry.
type ModelInfo struct {
	Model    string
	Provider Provider
	Encoding string
}

type modelEncoding struct {
	provider Provider
	encoding string
}

var (
	modelEncodingsMu    sync.RWMutex
	exactModelEncodings = map[string]modelEncoding{
		"o1":                     {ProviderOpenAI, EncodingO200KBase},
		"o3":                     {ProviderOpenAI, EncodingO200KBase},
		"o4-mini":                {ProviderOpenAI, EncodingO200KBase},
		"gpt-5":                  {ProviderOpenAI, EncodingO200KBase},
		"gpt-4.1":                {ProviderOpenAI, EncodingO200KBase},
		"gpt-4o":                 {ProviderOpenAI, EncodingO200KBase},
		"gpt-4":                  {ProviderOpenAI, EncodingCL100KBase},
		"gpt-3.5":                {ProviderOpenAI, EncodingCL100KBase},
		"gpt-3.5-turbo":          {ProviderOpenAI, EncodingCL100KBase},
		"gpt-35-turbo":           {ProviderOpenAI, EncodingCL100KBase},
		"davinci-002":            {ProviderOpenAI, EncodingCL100KBase},
		"babbage-002":            {ProviderOpenAI, EncodingCL100KBase},
		"text-embedding-ada-002": {ProviderOpenAI, EncodingCL100KBase},
		"text-embedding-3-small": {ProviderOpenAI, EncodingCL100KBase},
		"text-embedding-3-large": {ProviderOpenAI, EncodingCL100KBase},
	}

	prefixModelEncodings = []struct {
		prefix   string
		provider Provider
		encoding string
	}{
		{"o1-", ProviderOpenAI, EncodingO200KBase},
		{"o3-", ProviderOpenAI, EncodingO200KBase},
		{"o4-mini-", ProviderOpenAI, EncodingO200KBase},
		{"gpt-5-", ProviderOpenAI, EncodingO200KBase},
		{"gpt-4.5-", ProviderOpenAI, EncodingO200KBase},
		{"gpt-4.1-", ProviderOpenAI, EncodingO200KBase},
		{"chatgpt-4o-", ProviderOpenAI, EncodingO200KBase},
		{"gpt-4o-", ProviderOpenAI, EncodingO200KBase},
		{"gpt-4-", ProviderOpenAI, EncodingCL100KBase},
		{"gpt-3.5-turbo-", ProviderOpenAI, EncodingCL100KBase},
		{"gpt-35-turbo-", ProviderOpenAI, EncodingCL100KBase},
		{"gpt-oss-", ProviderOpenAI, EncodingO200KHarmony},
		{"ft:gpt-4o", ProviderOpenAI, EncodingO200KBase},
		{"ft:gpt-4", ProviderOpenAI, EncodingCL100KBase},
		{"ft:gpt-3.5-turbo", ProviderOpenAI, EncodingCL100KBase},
		{"ft:davinci-002", ProviderOpenAI, EncodingCL100KBase},
		{"ft:babbage-002", ProviderOpenAI, EncodingCL100KBase},
	}
)

var engineCache sync.Map

var engineBuildMu sync.Mutex

// EncodingFactory constructs a tokenizer engine for an encoding name.
type EncodingFactory func() (ModelEngine, error)

var (
	encodingFactoriesMu sync.RWMutex
	encodingFactories   = map[string]EncodingFactory{
		EncodingCL100KBase: func() (ModelEngine, error) {
			return newEngine(EncodingCL100KBase, cl100kBaseData, segmenterFunc(nextCL100K), cl100kSpecialTokens())
		},
		EncodingO200KBase: func() (ModelEngine, error) {
			return newEngine(EncodingO200KBase, o200kBaseData, segmenterFunc(nextO200K), o200kBaseSpecialTokens())
		},
		EncodingO200KHarmony: func() (ModelEngine, error) {
			return newEngine(EncodingO200KHarmony, o200kBaseData, segmenterFunc(nextO200K), o200kHarmonySpecialTokens())
		},
	}
)

// RegisterEncoding adds a custom tokenizer engine factory for a non-built-in encoding.
func RegisterEncoding(encoding string, factory EncodingFactory) error {
	if encoding == "" {
		return errors.New("omnitoken: encoding name is required")
	}
	if factory == nil {
		return errors.New("omnitoken: encoding factory is required")
	}

	encodingFactoriesMu.Lock()
	defer encodingFactoriesMu.Unlock()
	if _, exists := encodingFactories[encoding]; exists {
		return fmt.Errorf("omnitoken: encoding already registered: %s", encoding)
	}
	encodingFactories[encoding] = factory
	return nil
}

// RegisterModel maps an exact model name to a registered encoding.
func RegisterModel(model string, encoding string) error {
	return RegisterProviderModel(ProviderCustom, model, encoding)
}

// RegisterProviderModel maps an exact provider model name to a registered encoding.
func RegisterProviderModel(provider Provider, model string, encoding string) error {
	if model == "" {
		return errors.New("omnitoken: model name is required")
	}
	if provider == "" {
		return errors.New("omnitoken: provider is required")
	}
	if err := requireRegisteredEncoding(encoding); err != nil {
		return err
	}

	modelEncodingsMu.Lock()
	defer modelEncodingsMu.Unlock()
	if _, exists := exactModelEncodings[model]; exists {
		return fmt.Errorf("omnitoken: model already registered: %s", model)
	}
	exactModelEncodings[model] = modelEncoding{provider: provider, encoding: encoding}
	return nil
}

// RegisterModelPrefix maps a model-name prefix to a registered encoding.
func RegisterModelPrefix(prefix string, encoding string) error {
	return RegisterProviderModelPrefix(ProviderCustom, prefix, encoding)
}

// RegisterProviderModelPrefix maps a provider model-name prefix to a registered encoding.
func RegisterProviderModelPrefix(provider Provider, prefix string, encoding string) error {
	if prefix == "" {
		return errors.New("omnitoken: model prefix is required")
	}
	if provider == "" {
		return errors.New("omnitoken: provider is required")
	}
	if err := requireRegisteredEncoding(encoding); err != nil {
		return err
	}

	modelEncodingsMu.Lock()
	defer modelEncodingsMu.Unlock()
	for _, entry := range prefixModelEncodings {
		if entry.prefix == prefix {
			return fmt.Errorf("omnitoken: model prefix already registered: %s", prefix)
		}
	}
	prefixModelEncodings = append(prefixModelEncodings, struct {
		prefix   string
		provider Provider
		encoding string
	}{prefix: prefix, provider: provider, encoding: encoding})
	return nil
}

// ForModel resolves a model name to a tokenizer engine.
func ForModel(model string) (ModelEngine, error) {
	info, err := ResolveModel(model)
	if err != nil {
		return nil, err
	}
	return ForEncoding(info.Encoding)
}

// ResolveModel returns provider and encoding metadata for a model name.
func ResolveModel(model string) (ModelInfo, error) {
	mapping, ok := encodingForModel(model)
	if !ok {
		return ModelInfo{}, fmt.Errorf("%w: %s", ErrUnsupportedModel, model)
	}
	return ModelInfo{Model: model, Provider: mapping.provider, Encoding: mapping.encoding}, nil
}

// ForEncoding resolves an encoding name to a tokenizer engine.
func ForEncoding(encoding string) (ModelEngine, error) {
	if cached, ok := engineCache.Load(encoding); ok {
		return cached.(ModelEngine), nil
	}
	return buildEncodingOnce(encoding)
}

func buildEncodingOnce(encoding string) (ModelEngine, error) {
	engineBuildMu.Lock()
	defer engineBuildMu.Unlock()

	if cached, ok := engineCache.Load(encoding); ok {
		return cached.(ModelEngine), nil
	}

	engine, err := buildEncoding(encoding)
	if err != nil {
		return nil, err
	}
	actual, _ := engineCache.LoadOrStore(encoding, engine)
	return actual.(ModelEngine), nil
}

func buildEncoding(name string) (ModelEngine, error) {
	encodingFactoriesMu.RLock()
	factory := encodingFactories[name]
	encodingFactoriesMu.RUnlock()
	if factory == nil {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedEncoding, name)
	}
	return factory()
}

func requireRegisteredEncoding(encoding string) error {
	if encoding == "" {
		return errors.New("omnitoken: encoding name is required")
	}
	encodingFactoriesMu.RLock()
	_, ok := encodingFactories[encoding]
	encodingFactoriesMu.RUnlock()
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnsupportedEncoding, encoding)
	}
	return nil
}

func encodingForModel(model string) (modelEncoding, bool) {
	modelEncodingsMu.RLock()
	defer modelEncodingsMu.RUnlock()
	if mapping, ok := exactModelEncodings[model]; ok {
		return mapping, true
	}
	for _, entry := range prefixModelEncodings {
		if strings.HasPrefix(model, entry.prefix) {
			return modelEncoding{provider: entry.provider, encoding: entry.encoding}, true
		}
	}
	return modelEncoding{}, false
}
