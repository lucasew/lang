package ro

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Java RomanianTagger resource path: super("/ro/romanian.dict", new Locale("ro")).
const RomanianDictPath = "/ro/romanian.dict"

// Java package-private ctor path used by RomanianTaggerDiacriticsTest.
const RomanianTestDiacriticsDictPath = "/ro/test_diacritics.dict"

// RomanianTagger ports org.languagetool.tagging.ro.RomanianTagger
// (thin BaseTagger: super(dictPath, Locale("ro"))).
type RomanianTagger struct {
	*tagging.BaseTagger
}

// NewRomanianTagger builds a RomanianTagger over the given WordTagger.
// Java: super("/ro/romanian.dict", new Locale("ro")) → tagLowercaseWithUppercase true.
func NewRomanianTagger(wt tagging.WordTagger) *RomanianTagger {
	return NewRomanianTaggerWithPath(wt, RomanianDictPath)
}

// NewRomanianTaggerWithPath ports the package-private Java ctor
// RomanianTagger(String dictPath) used by diacritics tests.
func NewRomanianTaggerWithPath(wt tagging.WordTagger, dictPath string) *RomanianTagger {
	return &RomanianTagger{BaseTagger: tagging.NewBaseTagger(wt, dictPath, "ro", true)}
}

// Tag ports BaseTagger.tag via getAnalyzedTokens (RomanianTagger has no Java override).
func (t *RomanianTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
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
