package ca

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Process-wide CatalanTagger-equivalent for FindSuggestionsFilter.Tag
// (Java Catalan.getTagger() / CatalanTagger.INSTANCE).

var (
	caFindSugTagMu   sync.RWMutex
	caFindSugTagWord func(token string) []languagetool.TokenTag
)

// WireCatalanFilterTaggerFromTagWord installs lt.TagWord for FindSuggestions Tag path.
func WireCatalanFilterTaggerFromTagWord(tw func(token string) []languagetool.TokenTag) {
	caFindSugTagMu.Lock()
	defer caFindSugTagMu.Unlock()
	caFindSugTagWord = tw
}

// ClearCatalanFindSuggestionsTagger clears the process-wide tagger (tests).
func ClearCatalanFindSuggestionsTagger() {
	WireCatalanFilterTaggerFromTagWord(nil)
}

func getCatalanFindSugTagWord() func(token string) []languagetool.TokenTag {
	caFindSugTagMu.RLock()
	defer caFindSugTagMu.RUnlock()
	return caFindSugTagWord
}

// FilterTagWord ports getTagger().tag for FindSuggestions (fail-closed when unwired).
func FilterTagWord(word string) *languagetool.AnalyzedTokenReadings {
	tw := getCatalanFindSugTagWord()
	if tw == nil {
		return nil
	}
	if strings.TrimSpace(word) == "" {
		return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(word, nil, nil))
	}
	tags := tw(word)
	if len(tags) == 0 {
		return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(word, nil, nil))
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
		readings = append(readings, languagetool.NewAnalyzedToken(word, pos, lemma))
	}
	return languagetool.NewAnalyzedTokenReadingsList(readings, 0)
}
