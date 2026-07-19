package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const CatalanDictPath = "/ca/catalan.dict"

// CatalanTagger ports org.languagetool.tagging.ca.CatalanTagger.
type CatalanTagger struct {
	*tagging.BaseTagger
}

func NewCatalanTagger(wt tagging.WordTagger) *CatalanTagger {
	return &CatalanTagger{BaseTagger: tagging.NewBaseTagger(wt, CatalanDictPath, "ca", false)}
}

func (t *CatalanTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		// Java CatalanTagger: typewriter apostrophe hack + setTypographicApostrophe.
		w := word
		containsTypographicApostrophe := false
		if len(w) > 1 && strings.Contains(w, "’") {
			containsTypographicApostrophe = true
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
		atr := languagetool.NewAnalyzedTokenReadingsList(readings, pos)
		if containsTypographicApostrophe {
			atr.SetTypographicApostrophe(true)
		}
		out = append(out, atr)
		pos += len([]rune(word))
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
