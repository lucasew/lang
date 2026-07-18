package en

import (
	"strings"
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// english_filter_tagger wires official english.dict into FindSuggestionsFilter
// desiredPostag checks (Java AbstractFindSuggestionsFilter.getTagger()).

var (
	filterTagMu   sync.RWMutex
	filterTagWord func(token string) []languagetool.TokenTag
)

// WireEnglishFilterTagger opens CFSA2 english.dict for filter POS probes.
func WireEnglishFilterTagger(dictPath string) bool {
	if strings.TrimSpace(dictPath) == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	// Reuse BinaryEnglishTagWord case/apostrophe logic (Java EnglishTagger.tag).
	tw := BinaryEnglishTagWord(d)
	if tw == nil {
		return false
	}
	filterTagMu.Lock()
	filterTagWord = tw
	filterTagMu.Unlock()
	return true
}

// ClearEnglishFilterTagger clears the process-wide filter tagger (tests).
func ClearEnglishFilterTagger() {
	filterTagMu.Lock()
	filterTagWord = nil
	filterTagMu.Unlock()
}

func getFilterTagWord() func(string) []languagetool.TokenTag {
	filterTagMu.RLock()
	defer filterTagMu.RUnlock()
	return filterTagWord
}

// FilterTaggerAvailable reports whether a filter POS dict is wired.
func FilterTaggerAvailable() bool {
	return getFilterTagWord() != nil
}

// FilterSuggestionMatchesPostag tags suggestion with EnglishTagger-equivalent
// lookup and tests MatchesPosTagRegex(desiredPostag).
// Without a tagger, returns false (fail-closed: do not invent POS matches).
func FilterSuggestionMatchesPostag(suggestion, desiredPostag string) bool {
	if desiredPostag == "" {
		return false
	}
	tw := getFilterTagWord()
	if tw == nil {
		return false
	}
	tags := tw(suggestion)
	if len(tags) == 0 {
		return false
	}
	readings := make([]*languagetool.AnalyzedToken, 0, len(tags))
	for _, t := range tags {
		var pos, lemma *string
		if t.POS != "" {
			p := t.POS
			pos = &p
		}
		if t.Lemma != "" {
			l := t.Lemma
			lemma = &l
		}
		readings = append(readings, languagetool.NewAnalyzedToken(suggestion, pos, lemma))
	}
	atr := languagetool.NewAnalyzedTokenReadingsList(readings, 0)
	return atr.MatchesPosTagRegex(desiredPostag)
}

// FilterOriginalMatchesPostag tags the original surface (pre-disambig check in Java).
func FilterOriginalMatchesPostag(token, desiredPostag string) bool {
	return FilterSuggestionMatchesPostag(token, desiredPostag)
}
