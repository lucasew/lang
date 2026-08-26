package es

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Process-wide SpanishTagger-equivalent for FindSuggestionsFilter.Tag
// (Java Spanish.getTagger() / SpanishTagger.INSTANCE).

var (
	esFindSugTagMu   sync.RWMutex
	esFindSugTagWord func(token string) []languagetool.TokenTag
)

// WireSpanishFilterTaggerFromTagWord installs lt.TagWord for FindSuggestions Tag path.
func WireSpanishFilterTaggerFromTagWord(tw func(token string) []languagetool.TokenTag) {
	esFindSugTagMu.Lock()
	defer esFindSugTagMu.Unlock()
	esFindSugTagWord = tw
}

// ClearSpanishFindSuggestionsTagger clears the process-wide tagger (tests).
func ClearSpanishFindSuggestionsTagger() {
	WireSpanishFilterTaggerFromTagWord(nil)
}

func getSpanishFindSugTagWord() func(token string) []languagetool.TokenTag {
	esFindSugTagMu.RLock()
	defer esFindSugTagMu.RUnlock()
	return esFindSugTagWord
}

// FilterTagWord ports getTagger().tag for FindSuggestions (fail-closed when unwired).
func FilterTagWord(word string) *languagetool.AnalyzedTokenReadings {
	tw := getSpanishFindSugTagWord()
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
