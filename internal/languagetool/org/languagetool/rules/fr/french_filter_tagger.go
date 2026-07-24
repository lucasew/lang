package fr

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Process-wide FrenchTagger-equivalent for FindSuggestionsFilter / AbstractFindSuggestionsFilter.Tag
// (Java French.getInstance().getTagger()). Separate from PartialPosTagFilter POS-only list.

var (
	frFindSugTagMu   sync.RWMutex
	frFindSugTagWord func(token string) []languagetool.TokenTag
)

// WireFrenchFindSuggestionsTagger installs lt.TagWord for FindSuggestions Tag path.
func WireFrenchFindSuggestionsTagger(tw func(token string) []languagetool.TokenTag) {
	frFindSugTagMu.Lock()
	defer frFindSugTagMu.Unlock()
	frFindSugTagWord = tw
}

// ClearFrenchFindSuggestionsTagger clears the process-wide FindSuggestions tagger (tests).
func ClearFrenchFindSuggestionsTagger() {
	WireFrenchFindSuggestionsTagger(nil)
}

func getFrenchFindSugTagWord() func(token string) []languagetool.TokenTag {
	frFindSugTagMu.RLock()
	defer frFindSugTagMu.RUnlock()
	return frFindSugTagWord
}

// FilterTagWord ports getTagger().tag for FindSuggestions (fail-closed when unwired).
func FilterTagWord(word string) *languagetool.AnalyzedTokenReadings {
	tw := getFrenchFindSugTagWord()
	if tw == nil {
		return nil
	}
	if tools.JavaStringTrim(word) == "" {
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
