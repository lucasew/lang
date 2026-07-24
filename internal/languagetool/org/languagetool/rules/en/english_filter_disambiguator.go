package en

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// filterDisambiguator is the process-wide EN hybrid used by
// EnglishPartialPosTagFilter (Java Languages.getLanguageForShortCode("en").getDisambiguator()).
type filterDisambiguator interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

var (
	filterDisambigMu sync.RWMutex
	filterDisambig   filterDisambiguator
)

// WireEnglishFilterDisambiguator installs the EN hybrid for grammar filter probes
// that re-tag + disambiguate a single token (EnglishPartialPosTagFilter).
// Call from RegisterEnglishHybridDisambiguator when the hybrid is ready.
func WireEnglishFilterDisambiguator(d filterDisambiguator) {
	filterDisambigMu.Lock()
	filterDisambig = d
	filterDisambigMu.Unlock()
}

// ClearEnglishFilterDisambiguator clears the process-wide filter disambiguator (tests).
func ClearEnglishFilterDisambiguator() {
	filterDisambigMu.Lock()
	filterDisambig = nil
	filterDisambigMu.Unlock()
}

func getFilterDisambiguator() filterDisambiguator {
	filterDisambigMu.RLock()
	defer filterDisambigMu.RUnlock()
	return filterDisambig
}
