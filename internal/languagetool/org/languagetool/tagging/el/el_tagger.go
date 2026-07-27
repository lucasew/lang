package el

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// GreekDictPath ports Java GreekTagger resource: super("/el/greek.dict", Locale("el")).
const GreekDictPath = "/el/greek.dict"

// GreekTagger ports org.languagetool.tagging.el.GreekTagger
// (BaseTagger + additionalTags via GreekAnalyzer — analyzer twin is a leaf gap).
type GreekTagger struct {
	*tagging.BaseTagger
}

// NewGreekTagger builds a GreekTagger over the given WordTagger.
// Java: super("/el/greek.dict", new Locale("el")) → tagLowercaseWithUppercase true.
func NewGreekTagger(wt tagging.WordTagger) *GreekTagger {
	return &GreekTagger{BaseTagger: tagging.NewBaseTagger(wt, GreekDictPath, "el", true)}
}

func (t *GreekTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		w := strings.ReplaceAll(word, "’", "'")
		var readings []*languagetool.AnalyzedToken
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
		pos += tagging.UTF16Len(word)
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
