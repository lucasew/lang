package ml

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Java MalayalamTagger resource path: super("/ml/malayalam.dict", new Locale("ml")).
const MalayalamDictPath = "/ml/malayalam.dict"

// MalayalamTagger ports org.languagetool.tagging.ml.MalayalamTagger
// (thin BaseTagger: super("/ml/malayalam.dict", Locale("ml"))).
type MalayalamTagger struct {
	*tagging.BaseTagger
}

// NewMalayalamTagger builds a MalayalamTagger over the given WordTagger.
// Java: super("/ml/malayalam.dict", new Locale("ml")) → tagLowercaseWithUppercase true.
func NewMalayalamTagger(wt tagging.WordTagger) *MalayalamTagger {
	return &MalayalamTagger{BaseTagger: tagging.NewBaseTagger(wt, MalayalamDictPath, "ml", true)}
}

// Tag ports BaseTagger.tag via getAnalyzedTokens (MalayalamTagger has no Java override).
func (t *MalayalamTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
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
