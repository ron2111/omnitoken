package omnitoken

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

var registryTestCounter uint64

func TestForEncodingBuildsCustomEncodingOnce(t *testing.T) {
	encoding := fmt.Sprintf("test_singleflight_%d", atomic.AddUint64(&registryTestCounter, 1))
	started := make(chan struct{})
	block := make(chan struct{})
	var builds atomic.Int64

	if err := RegisterEncoding(encoding, func() (ModelEngine, error) {
		builds.Add(1)
		select {
		case <-started:
		default:
			close(started)
		}
		<-block
		return fixedCountEngine(7), nil
	}); err != nil {
		t.Fatal(err)
	}

	const goroutines = 32
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			engine, err := ForEncoding(encoding)
			if err != nil {
				errs <- err
				return
			}
			if got := engine.CountTokens("ignored"); got != 7 {
				errs <- fmt.Errorf("CountTokens = %d, want 7", got)
			}
		}()
	}

	<-started
	close(block)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := builds.Load(); got != 1 {
		t.Fatalf("factory builds = %d, want 1", got)
	}
}

func TestRegisteredRegistryListings(t *testing.T) {
	encoding := fmt.Sprintf("test_listing_%d", atomic.AddUint64(&registryTestCounter, 1))
	model := fmt.Sprintf("test-listing-model-%d", atomic.AddUint64(&registryTestCounter, 1))
	prefix := fmt.Sprintf("test-listing-prefix-%d-", atomic.AddUint64(&registryTestCounter, 1))
	if err := RegisterEncoding(encoding, func() (ModelEngine, error) { return fixedCountEngine(1), nil }); err != nil {
		t.Fatal(err)
	}
	if err := RegisterProviderModel(ProviderCustom, model, encoding); err != nil {
		t.Fatal(err)
	}
	if err := RegisterProviderModelPrefix(ProviderCustom, prefix, encoding); err != nil {
		t.Fatal(err)
	}

	if !containsString(RegisteredEncodings(), encoding) {
		t.Fatalf("RegisteredEncodings missing %q", encoding)
	}
	if !containsModel(RegisteredModels(), model, ProviderCustom, encoding) {
		t.Fatalf("RegisteredModels missing %q", model)
	}
	if !containsPrefix(RegisteredModelPrefixes(), prefix, ProviderCustom, encoding) {
		t.Fatalf("RegisteredModelPrefixes missing %q", prefix)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsModel(models []ModelInfo, model string, provider Provider, encoding string) bool {
	for _, info := range models {
		if info.Model == model && info.Provider == provider && info.Encoding == encoding {
			return true
		}
	}
	return false
}

func containsPrefix(prefixes []ModelPrefixInfo, prefix string, provider Provider, encoding string) bool {
	for _, info := range prefixes {
		if info.Prefix == prefix && info.Provider == provider && info.Encoding == encoding {
			return true
		}
	}
	return false
}
