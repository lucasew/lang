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

// NewMultiThreadedJLanguageTool ports MultiThreadedJLanguageTool(Language, int threadPoolSize).
func NewMultiThreadedJLanguageTool(languageCode string, threadPoolSize int) *MultiThreadedJLanguageTool {
	return NewMultiThreadedJLanguageToolFull(languageCode, "", threadPoolSize, nil)
}

// NewMultiThreadedJLanguageToolFull ports constructors with mother tongue and user config.
// threadPoolSize <= 0 uses availableProcessors (Java getDefaultThreadCount).
func NewMultiThreadedJLanguageToolFull(languageCode, motherTongue string, threadPoolSize int, userConfig *UserConfig) *MultiThreadedJLanguageTool {
	if threadPoolSize <= 0 {
		threadPoolSize = defaultThreadCount()
	}
	lt := NewJLanguageTool(languageCode)
	if motherTongue != "" {
		// mother tongue stored when field exists; Mode/UserConfig surface
		_ = motherTongue
	}
	if userConfig != nil {
		lt.UserConfig = userConfig
	}
	return &MultiThreadedJLanguageTool{
		JLanguageTool: lt,
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

// GetThreadPoolSize ports getThreadPoolSize.
func (lt *MultiThreadedJLanguageTool) GetThreadPoolSize() int {
	if lt == nil {
		return 0
	}
	return lt.poolSize
}

// AnalyzeSentenceCallable ports private AnalyzeSentenceCallable — analyze one sentence.
type AnalyzeSentenceCallable struct {
	LT       *MultiThreadedJLanguageTool
	Sentence string
}

func (c AnalyzeSentenceCallable) Call() (*AnalyzedSentence, error) {
	if c.LT == nil || c.LT.JLanguageTool == nil {
		return nil, nil
	}
	// Java: getAnalyzedSentence(sentence)
	sents := c.LT.Analyze(c.Sentence)
	if len(sents) == 0 {
		return NewAnalyzedSentence(nil), nil
	}
	return sents[0], nil
}

// ParagraphEndAnalyzeSentenceCallable ports ParagraphEndAnalyzeSentenceCallable.
type ParagraphEndAnalyzeSentenceCallable struct {
	AnalyzeSentenceCallable
}

func (c ParagraphEndAnalyzeSentenceCallable) Call() (*AnalyzedSentence, error) {
	s, err := c.AnalyzeSentenceCallable.Call()
	if err != nil || s == nil {
		return s, err
	}
	return MarkAsParagraphEnd(s), nil
}

// MarkAsParagraphEnd sets paragraph-end on the last token (Java markAsParagraphEnd).
func MarkAsParagraphEnd(s *AnalyzedSentence) *AnalyzedSentence {
	if s == nil {
		return nil
	}
	toks := s.GetTokens()
	if len(toks) == 0 {
		return s
	}
	// clone shallow tokens slice and mark last non-nil
	out := make([]*AnalyzedTokenReadings, len(toks))
	copy(out, toks)
	last := out[len(out)-1]
	if last != nil {
		last.SetParagraphEnd()
	}
	return NewAnalyzedSentence(out)
}

// AnalyzeSentencesParallel ports analyzeSentences override for multi-sentence input.
// Last sentence uses ParagraphEndAnalyzeSentenceCallable.
func (lt *MultiThreadedJLanguageTool) AnalyzeSentencesParallel(sentences []string) []*AnalyzedSentence {
	if lt == nil || lt.shutdown {
		return nil
	}
	if len(sentences) < 2 {
		var out []*AnalyzedSentence
		for _, s := range sentences {
			a, _ := AnalyzeSentenceCallable{LT: lt, Sentence: s}.Call()
			out = append(out, a)
		}
		return out
	}
	workers := lt.poolSize
	if workers > len(sentences) {
		workers = len(sentences)
	}
	if workers < 1 {
		workers = 1
	}
	type job struct {
		i int
		s string
	}
	jobs := make(chan job, len(sentences))
	results := make([]*AnalyzedSentence, len(sentences))
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				var a *AnalyzedSentence
				if j.i == len(sentences)-1 {
					a, _ = ParagraphEndAnalyzeSentenceCallable{AnalyzeSentenceCallable{LT: lt, Sentence: j.s}}.Call()
				} else {
					a, _ = AnalyzeSentenceCallable{LT: lt, Sentence: j.s}.Call()
				}
				results[j.i] = a
			}
		}()
	}
	for i, s := range sentences {
		jobs <- job{i: i, s: s}
	}
	close(jobs)
	wg.Wait()
	return results
}

// CheckSentences runs matchers over sentences in parallel (up to poolSize workers).
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

// Shutdown ports shutdown() — marks closed; further CheckSentences panics.
func (lt *MultiThreadedJLanguageTool) Shutdown() {
	if lt != nil {
		lt.shutdown = true
	}
}

// ShutdownWhenDone ports shutdownWhenDone (graceful; Go model same as Shutdown).
func (lt *MultiThreadedJLanguageTool) ShutdownWhenDone() {
	lt.Shutdown()
}

func (lt *MultiThreadedJLanguageTool) IsShutdown() bool {
	return lt != nil && lt.shutdown
}

// Check runs sentence checkers in parallel across sentences, then merges document-offset matches.
// Text-level rules run sequentially after the parallel sentence pass.
func (lt *MultiThreadedJLanguageTool) Check(text string) []LocalMatch {
	if lt == nil || lt.shutdown {
		return nil
	}
	if lt.Cancelled != nil && lt.Cancelled() {
		return nil
	}
	if lt.poolSize <= 1 {
		return lt.JLanguageTool.Check(text)
	}
	sents := lt.Analyze(text)
	if len(sents) <= 1 {
		return lt.JLanguageTool.Check(text)
	}

	runSentence := lt.Mode != ModeTextLevelOnly && lt.paraModeOrNormal() != ParagraphOnlyPara
	runTextLevel := lt.Mode != ModeAllButTextLevel && lt.paraModeOrNormal() != ParagraphOnlyNonPara
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

		// Java TextCheckCallable: computeSentenceData + adjustRuleMatchPos per sentence.
		data := sentenceDataFromAnalyzed(sents)
		for i, sd := range data {
			s := sd.Analyzed
			if s == nil {
				continue
			}
			if lt.ListUnknownWords {
				lt.collectUnknown(s)
			}
			sentTypoApos := sentenceHasTypographicApostrophe(s)
			for _, m := range results[i].matches {
				adapted := remapLocalMatchToDocument(m, sd, sentTypoApos)
				lt.notifyMatchFound(adapted)
				out = append(out, adapted)
			}
		}
	} else if lt.ListUnknownWords {
		for _, s := range sents {
			lt.collectUnknown(s)
		}
	}

	if runTextLevel && len(lt.textLevelCheckers) > 0 {
		if lt.Cancelled == nil || !lt.Cancelled() {
			// Reuse sentence data for text-level line/column (same as single-thread Check).
			tlData := sentenceDataFromAnalyzed(sents)
			for _, tc := range lt.textLevelCheckers {
				if tc.id != "" && lt.isRuleDisabled(tc.id) {
					continue
				}
				for _, m := range tc.fn(sents) {
					adapted := AdaptTextLevelLocalMatch(m, tlData, nil)
					lt.notifyMatchFound(adapted)
					out = append(out, adapted)
				}
			}
		}
	}
	out = lt.applyRulePriorities(out)
	out = CleanSameRuleGroupLocalMatches(out)
	if en := lt.enabledRulesForFilters(); len(en) > 0 {
		for i := range out {
			out[i].EnabledRules = en
		}
	}
	if lt.FilterRuleMatches != nil {
		out = lt.FilterRuleMatches(out)
	}
	if !lt.disableCleanOverlapping {
		hidePremium := lt.UserConfig != nil && lt.UserConfig.HidePremiumMatches
		out = CleanOverlappingLocalMatchesOpts(out, CleanOverlapOpts{HidePremiumMatches: hidePremium})
	}
	if lt.FilterRuleMatchesAfterOverlapping != nil {
		out = lt.FilterRuleMatchesAfterOverlapping(out)
	}
	return lt.filterMatchesByIgnore(text, out)
}
