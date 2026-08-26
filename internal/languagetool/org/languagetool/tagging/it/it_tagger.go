package it

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

const ItalianDictPath = "/it/italian.dict"

// ItalianTagger ports org.languagetool.tagging.it.ItalianTagger
// (thin BaseTagger: super("/it/italian.dict", Locale.ITALIAN)).
type ItalianTagger struct {
	*tagging.BaseTagger
}

// NewItalianTagger builds an ItalianTagger over the given WordTagger.
// Java: super("/it/italian.dict", Locale.ITALIAN) → tagLowercaseWithUppercase true.
func NewItalianTagger(wt tagging.WordTagger) *ItalianTagger {
	return &ItalianTagger{BaseTagger: tagging.NewBaseTagger(wt, ItalianDictPath, "it", true)}
}

// Tag ports BaseTagger.tag via getAnalyzedTokens (ItalianTagger has no Java override).
func (t *ItalianTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
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
