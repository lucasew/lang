package fr

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const FrenchDictPath = "/fr/french.dict"

// FrenchTagger ports org.languagetool.tagging.fr.FrenchTagger.
type FrenchTagger struct {
	*tagging.BaseTagger
}

func NewFrenchTagger(wt tagging.WordTagger) *FrenchTagger {
	return &FrenchTagger{BaseTagger: tagging.NewBaseTagger(wt, FrenchDictPath, "fr", false)}
}

func (t *FrenchTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		// Java FrenchTagger: apostrophe chunk tags (not setTypographicApostrophe).
		w := word
		containsTypewriterApostrophe := false
		containsTypographicApostrophe := false
		if len(w) > 1 {
			if strings.Contains(w, "'") {
				containsTypewriterApostrophe = true
			}
			if strings.Contains(w, "’") {
				containsTypographicApostrophe = true
				w = strings.ReplaceAll(w, "’", "'")
			}
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
		// Java: setChunkTags replaces list; typographic overwrites typewriter when both.
		if containsTypewriterApostrophe {
			atr.SetChunkTags([]string{"containsTypewriterApostrophe"})
		}
		if containsTypographicApostrophe {
			atr.SetChunkTags([]string{"containsTypographicApostrophe"})
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
