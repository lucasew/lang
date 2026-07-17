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
	shutdown bool
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
	if lt != nil && lt.shutdown {
		panic("MultiThreadedJLanguageTool has been shut down")
	}
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
// Shutdown marks the tool closed; further CheckSentences panics (Java IllegalStateException).
func (lt *MultiThreadedJLanguageTool) Shutdown() {
	if lt != nil {
		lt.shutdown = true
	}
}

func (lt *MultiThreadedJLanguageTool) IsShutdown() bool {
	return lt != nil && lt.shutdown
}

// Check runs sentence checkers in parallel across sentences, then merges document-offset matches.
// Text-level rules run sequentially after the parallel sentence pass (document-relative offsets).
// Falls back to sequential JLanguageTool.Check when pool size is 1 or a single sentence.
func (lt *MultiThreadedJLanguageTool) Check(text string) []LocalMatch {
	if lt == nil || lt.shutdown {
		return nil
	}
	if lt.Cancelled != nil && lt.Cancelled() {
		return nil
	}
	// sequential path is fine for small inputs
	if lt.poolSize <= 1 {
		return lt.JLanguageTool.Check(text)
	}
	sents := lt.Analyze(text)
	if len(sents) <= 1 {
		return lt.JLanguageTool.Check(text)
	}

	runSentence := lt.Mode != ModeTextLevelOnly
	runTextLevel := lt.Mode != ModeAllButTextLevel
	lt.unknown = map[string]struct{}{}

	type sentResult struct {
		idx     int
		matches []LocalMatch
	}
	results := make([]sentResult, len(sents))
	var out []LocalMatch

	if runSentence {
		jobs := make(chan int, len(sents))
		var wg sync.WaitGroup
		workers := lt.poolSize
		if workers > len(sents) {
			workers = len(sents)
		}
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range jobs {
					s := sents[i]
					var local []LocalMatch
					for _, c := range lt.checkers {
						if c == nil {
							continue
						}
						local = append(local, c(s)...)
					}
					results[i] = sentResult{idx: i, matches: local}
				}
			}()
		}
		for i := range sents {
			jobs <- i
		}
		close(jobs)
		wg.Wait()

		// map to document offsets
		srcRunes := []rune(text)
		searchFrom := 0
		for i, s := range sents {
			if s == nil {
				continue
			}
			if lt.ListUnknownWords {
				lt.collectUnknown(s)
			}
			stext := s.GetText()
			docBase := indexRunesFrom(srcRunes, []rune(stext), searchFrom)
			if docBase < 0 {
				docBase = searchFrom
			}
			for _, m := range results[i].matches {
				m.FromPos += docBase
				m.ToPos += docBase
				out = append(out, m)
			}
			searchFrom = docBase + len([]rune(stext))
		}
	} else if lt.ListUnknownWords {
		for _, s := range sents {
			lt.collectUnknown(s)
		}
	}

	if runTextLevel && len(lt.textLevelCheckers) > 0 {
		if lt.Cancelled == nil || !lt.Cancelled() {
			for _, tc := range lt.textLevelCheckers {
				if tc.id != "" && lt.isRuleDisabled(tc.id) {
					continue
				}
				out = append(out, tc.fn(sents)...)
			}
		}
	}
	return lt.filterMatchesByIgnore(text, out)
}
