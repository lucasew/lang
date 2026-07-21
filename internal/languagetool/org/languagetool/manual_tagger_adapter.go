package languagetool

import (
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// ManualTaggerAdapter ports org.languagetool.tokenizers.ManualTaggerAdapter
// (Java test helper: ManualTagger → Tagger).
//
// Lives in package languagetool (not tokenizers) to avoid import cycle:
// languagetool → tokenizers → languagetool (AnalyzedToken*).
type ManualTaggerAdapter struct {
	manual *tagging.ManualTagger
}

func NewManualTaggerAdapter(manual *tagging.ManualTagger) *ManualTaggerAdapter {
	return &ManualTaggerAdapter{manual: manual}
}

// Tag ports ManualTaggerAdapter.tag: lowercases for dict lookup, keeps surface token.
func (a *ManualTaggerAdapter) Tag(sentenceTokens []string) []*AnalyzedTokenReadings {
	if a == nil {
		return nil
	}
	tokenReadings := make([]*AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		var l []*AnalyzedToken
		if a.manual != nil {
			for _, tw := range a.manual.Tag(strings.ToLower(word)) {
				lemma := tw.GetLemma()
				tag := tw.GetPosTag()
				l = append(l, NewAnalyzedToken(word, &tag, &lemma))
			}
		}
		if len(l) == 0 {
			l = append(l, NewAnalyzedToken(word, nil, nil))
		}
		tokenReadings = append(tokenReadings, NewAnalyzedTokenReadingsList(l, pos))
		// Java: pos += word.length() — UTF-16 code units
		pos += len(utf16.Encode([]rune(word)))
	}
	return tokenReadings
}

// CreateNullToken ports ManualTaggerAdapter.createNullToken.
func (a *ManualTaggerAdapter) CreateNullToken(token string, startPos int) *AnalyzedTokenReadings {
	return NewAnalyzedTokenReadingsList(
		[]*AnalyzedToken{NewAnalyzedToken(token, nil, nil)},
		startPos,
	)
}

// CreateToken ports ManualTaggerAdapter.createToken (lemma null).
func (a *ManualTaggerAdapter) CreateToken(token, posTag string) *AnalyzedToken {
	var p *string
	if posTag != "" {
		p = &posTag
	}
	return NewAnalyzedToken(token, p, nil)
}
