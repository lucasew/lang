package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"strings"
)

const UkrainianDictPath = "/uk/uk.dict"

type UkrainianTagger struct{ *tagging.BaseTagger }

func NewUkrainianTagger(wt tagging.WordTagger) *UkrainianTagger {
	return &UkrainianTagger{BaseTagger: tagging.NewBaseTagger(wt, UkrainianDictPath, "uk", false)}
}

func (t *UkrainianTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		w := strings.ReplaceAll(word, "’", "'")
		var readings []*languagetool.AnalyzedToken
		if sp := SpecialPOSTag(w); sp != "" {
			p := sp
			lemma := w
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &lemma)}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		for _, tw := range t.TagWord(w) {
			readings = append(readings, toTok(word, tw))
		}
		lower := strings.ToLower(w)
		if len(readings) == 0 && w != lower && !tools.IsMixedCase(w) {
			for _, tw := range t.TagWord(lower) {
				readings = append(readings, toTok(word, tw))
			}
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
		}
		out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
		pos += len([]rune(word))
	}
	return out
}

func toTok(surface string, tw tagging.TaggedWord) *languagetool.AnalyzedToken {
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
