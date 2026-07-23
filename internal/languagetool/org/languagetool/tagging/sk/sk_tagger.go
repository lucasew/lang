package sk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Java SlovakTagger resource path: super("/sk/slovak.dict", new Locale("sk")).
const SlovakDictPath = "/sk/slovak.dict"

// SlovakTagger ports org.languagetool.tagging.sk.SlovakTagger
// (thin BaseTagger: super("/sk/slovak.dict", Locale("sk"))).
type SlovakTagger struct {
	*tagging.BaseTagger
}

// NewSlovakTagger builds a SlovakTagger over the given WordTagger.
// Java: super("/sk/slovak.dict", new Locale("sk")) → tagLowercaseWithUppercase true.
func NewSlovakTagger(wt tagging.WordTagger) *SlovakTagger {
	return &SlovakTagger{BaseTagger: tagging.NewBaseTagger(wt, SlovakDictPath, "sk", true)}
}

// Tag ports BaseTagger.tag via getAnalyzedTokens (SlovakTagger has no Java override).
func (t *SlovakTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
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
