package da

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Java DanishTagger resource path: super("/da/danish.dict", new Locale("da")).
const DanishDictPath = "/da/danish.dict"

// DanishTagger ports org.languagetool.tagging.da.DanishTagger
// (thin BaseTagger: super("/da/danish.dict", Locale("da"))).
type DanishTagger struct {
	*tagging.BaseTagger
}

// NewDanishTagger builds a DanishTagger over the given WordTagger.
// Java: super("/da/danish.dict", new Locale("da")) → tagLowercaseWithUppercase true.
func NewDanishTagger(wt tagging.WordTagger) *DanishTagger {
	return &DanishTagger{BaseTagger: tagging.NewBaseTagger(wt, DanishDictPath, "da", true)}
}

// Tag ports BaseTagger.tag via getAnalyzedTokens (DanishTagger has no Java override).
func (t *DanishTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
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
