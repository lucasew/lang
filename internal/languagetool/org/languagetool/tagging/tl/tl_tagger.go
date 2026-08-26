package tl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Java TagalogTagger resource path: super("/tl/tagalog.dict", new Locale("tl")).
const TagalogDictPath = "/tl/tagalog.dict"

// TagalogTagger ports org.languagetool.tagging.tl.TagalogTagger
// (thin BaseTagger: super("/tl/tagalog.dict", Locale("tl"))).
type TagalogTagger struct {
	*tagging.BaseTagger
}

// NewTagalogTagger builds a TagalogTagger over the given WordTagger.
// Java: super("/tl/tagalog.dict", new Locale("tl")) → tagLowercaseWithUppercase true.
func NewTagalogTagger(wt tagging.WordTagger) *TagalogTagger {
	return &TagalogTagger{BaseTagger: tagging.NewBaseTagger(wt, TagalogDictPath, "tl", true)}
}

// Tag ports BaseTagger.tag via getAnalyzedTokens (TagalogTagger has no Java override).
func (t *TagalogTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
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
