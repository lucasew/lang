package ta

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Java TamilTagger resource path: super("/ta/tamil.dict", new Locale("ta")).
const TamilDictPath = "/ta/tamil.dict"

// TamilTagger ports org.languagetool.language.tagging.TamilTagger
// (thin BaseTagger: super("/ta/tamil.dict", Locale("ta"))).
// Note: Java lives under language.tagging, not tagging.ta.
type TamilTagger struct {
	*tagging.BaseTagger
}

// NewTamilTagger builds a TamilTagger over the given WordTagger.
// Java: super("/ta/tamil.dict", new Locale("ta")) → tagLowercaseWithUppercase true.
func NewTamilTagger(wt tagging.WordTagger) *TamilTagger {
	return &TamilTagger{BaseTagger: tagging.NewBaseTagger(wt, TamilDictPath, "ta", true)}
}

// Tag ports BaseTagger.tag via getAnalyzedTokens (TamilTagger has no Java override).
func (t *TamilTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
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
