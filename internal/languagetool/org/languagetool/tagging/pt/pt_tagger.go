package pt

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const PortugueseDictPath = "/pt/portuguese.dict"

// PortugueseTagger ports org.languagetool.tagging.pt.PortugueseTagger.
type PortugueseTagger struct {
	*tagging.BaseTagger
}

func NewPortugueseTagger(wt tagging.WordTagger) *PortugueseTagger {
	return &PortugueseTagger{BaseTagger: tagging.NewBaseTagger(wt, PortugueseDictPath, "pt", false)}
}

func (t *PortugueseTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
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
		if len(readings) == 0 {
			for _, cr := range ContractionReadings(w) {
				pos, lemma := cr.POS, cr.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &pos, &lemma))
			}
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
