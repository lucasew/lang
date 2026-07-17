package languagetool

import (
	"fmt"
	"sync"
	"testing"
)

// ConcurrencyAnalyzeSmoke runs parallel Analyze calls (Java AbstractLanguageConcurrencyTest soft).
// Full spell-checker concurrent failure matrix deferred; this asserts Analyze is race-free.
func ConcurrencyAnalyzeSmoke(t *testing.T, langCode, sample string) {
	t.Helper()
	if sample == "" {
		sample = "test"
	}
	lt := NewJLanguageTool(langCode)
	const threads = 8
	const runs = 5
	var wg sync.WaitGroup
	errCh := make(chan error, threads)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := 0; r < runs; r++ {
				sents := lt.Analyze(sample)
				if len(sents) == 0 {
					errCh <- fmt.Errorf("%s: empty analyze for %q", langCode, sample)
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Error(err)
		}
	}
}
