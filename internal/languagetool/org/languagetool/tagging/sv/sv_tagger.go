package sv

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Java SwedishTagger resource path: super("/sv/swedish.dict", new Locale("sv")).
const SwedishDictPath = "/sv/swedish.dict"

// SwedishTagger ports org.languagetool.tagging.sv.SwedishTagger
// (thin BaseTagger: super("/sv/swedish.dict", Locale("sv"))).
type SwedishTagger struct {
	*tagging.BaseTagger
}

// NewSwedishTagger builds a SwedishTagger over the given WordTagger.
// Java: super("/sv/swedish.dict", new Locale("sv")) → tagLowercaseWithUppercase true.
func NewSwedishTagger(wt tagging.WordTagger) *SwedishTagger {
	return &SwedishTagger{BaseTagger: tagging.NewBaseTagger(wt, SwedishDictPath, "sv", true)}
}

// Tag ports BaseTagger.tag via getAnalyzedTokens (SwedishTagger has no Java override).
func (t *SwedishTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		var readings []*languagetool.AnalyzedToken
		// Java getAnalyzedTokens(word) — BaseTagger case-merge via TagWord.
		for _, tw := range t.TagWord(word) {
			readings = append(readings, tagged(word, tw))
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
		}
		out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
		// Java: pos += word.length() (UTF-16 code units).
		pos += tagging.UTF16Len(word)
	}
	return out
}

func tagged(surface string, tw tagging.TaggedWord) *languagetool.AnalyzedToken {
	var pos, lemma *string
	if tw.PosTag != "" {
		p := tw.PosTag
		pos = &p
	}
	if tw.Lemma != "" {
		l := tw.Lemma
		lemma = &l
	}
	return languagetool.NewAnalyzedToken(surface, pos, lemma)
}
