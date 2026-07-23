package km

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Java KhmerTagger resource path: super("/km/khmer.dict", new Locale("km")).
const KhmerDictPath = "/km/khmer.dict"

// KhmerTagger ports org.languagetool.tagging.km.KhmerTagger
// (thin BaseTagger: super("/km/khmer.dict", Locale("km"))).
type KhmerTagger struct {
	*tagging.BaseTagger
}

// NewKhmerTagger builds a KhmerTagger over the given WordTagger.
// Java: super("/km/khmer.dict", new Locale("km")) → tagLowercaseWithUppercase true.
func NewKhmerTagger(wt tagging.WordTagger) *KhmerTagger {
	return &KhmerTagger{BaseTagger: tagging.NewBaseTagger(wt, KhmerDictPath, "km", true)}
}

// Tag ports BaseTagger.tag via getAnalyzedTokens (KhmerTagger has no Java override).
func (t *KhmerTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
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
