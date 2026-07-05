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
