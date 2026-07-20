package it

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const ItalianDictPath = "/it/italian.dict"

// ItalianTagger ports org.languagetool.tagging.it.ItalianTagger.
type ItalianTagger struct {
	*tagging.BaseTagger
}

func NewItalianTagger(wt tagging.WordTagger) *ItalianTagger {
	return &ItalianTagger{BaseTagger: tagging.NewBaseTagger(wt, ItalianDictPath, "it", false)}
}

func (t *ItalianTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		w := word
		if strings.Contains(w, "’") {
			w = strings.ReplaceAll(w, "’", "'")
		}
		var readings []*languagetool.AnalyzedToken
		for _, tw := range t.TagWord(w) {
			readings = append(readings, tagged(word, tw))
		}
		lower := strings.ToLower(w)
		if len(readings) == 0 && w != lower && !tools.IsMixedCase(w) {
			for _, tw := range t.TagWord(lower) {
				readings = append(readings, tagged(word, tw))
			}
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
		}
		out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
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
