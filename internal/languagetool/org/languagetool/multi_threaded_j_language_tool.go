package languagetool

import (
	"runtime"
	"sync"
)

// SentenceMatcherFunc matches one analyzed sentence (package-local surface
// without importing rules to avoid cycles). Implementations wrap pattern matchers.
type SentenceMatcherFunc func(sentence *AnalyzedSentence) error

// MultiThreadedJLanguageTool ports org.languagetool.MultiThreadedJLanguageTool
// as a parallel sentence checker over registered matchers.
// Not safe for concurrent use of a single instance (like Java).
type MultiThreadedJLanguageTool struct {
	*JLanguageTool
	poolSize int
	// Matchers run for each sentence; errors abort that sentence's remaining matchers.
	Matchers []SentenceMatcherFunc
}

func NewMultiThreadedJLanguageTool(languageCode string, threadPoolSize int) *MultiThreadedJLanguageTool {
	if threadPoolSize <= 0 {
		threadPoolSize = defaultThreadCount()
	}
	return &MultiThreadedJLanguageTool{
		JLanguageTool: NewJLanguageTool(languageCode),
		poolSize:      threadPoolSize,
	}
}

func defaultThreadCount() int {
	n := runtime.NumCPU()
	if n < 1 {
		return 1
	}
	return n
}

func (lt *MultiThreadedJLanguageTool) GetThreadPoolSize() int { return lt.poolSize }

// CheckSentences runs matchers over sentences in parallel (up to poolSize workers).
// Order of results is not guaranteed; per-sentence matcher order is preserved.
func (lt *MultiThreadedJLanguageTool) CheckSentences(sentences []*AnalyzedSentence) error {
	if len(sentences) == 0 || len(lt.Matchers) == 0 {
		return nil
	}
	workers := lt.poolSize
	if workers > len(sentences) {
		workers = len(sentences)
	}
	if workers < 1 {
		workers = 1
	}
	jobs := make(chan *AnalyzedSentence, len(sentences))
	errCh := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for s := range jobs {
				for _, m := range lt.Matchers {
					if m == nil {
						continue
					}
					if err := m(s); err != nil {
						select {
						case errCh <- err:
						default:
						}
						return
					}
				}
			}
		}()
	}
	for _, s := range sentences {
		jobs <- s
	}
	close(jobs)
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

// Shutdown is a no-op for the Go worker model (goroutines exit with CheckSentences).
func (lt *MultiThreadedJLanguageTool) Shutdown() {}
