package languagetool

import (
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// ManualTaggerAdapter ports the Java test helper ManualTaggerAdapter.
type ManualTaggerAdapter struct {
	manual *tagging.ManualTagger
}

func NewManualTaggerAdapter(manual *tagging.ManualTagger) *ManualTaggerAdapter {
	return &ManualTaggerAdapter{manual: manual}
}

func (a *ManualTaggerAdapter) Tag(sentenceTokens []string) []*AnalyzedTokenReadings {
	var tokenReadings []*AnalyzedTokenReadings
	pos := 0
	for _, word := range sentenceTokens {
		var l []*AnalyzedToken
		for _, tw := range a.manual.Tag(strings.ToLower(word)) {
			lemma := tw.GetLemma()
			tag := tw.GetPosTag()
			l = append(l, NewAnalyzedToken(word, &tag, &lemma))
		}
		if len(l) == 0 {
			l = append(l, NewAnalyzedToken(word, nil, nil))
		}
		tokenReadings = append(tokenReadings, NewAnalyzedTokenReadingsList(l, pos))
		pos += len(utf16.Encode([]rune(word)))
	}
	return tokenReadings
}
