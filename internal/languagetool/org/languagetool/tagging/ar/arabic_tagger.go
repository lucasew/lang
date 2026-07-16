package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const ArabicDictPath = "/ar/arabic.dict"

// ArabicTagger ports org.languagetool.tagging.ar.ArabicTagger (dict lookup without Morfologik prefixes).
type ArabicTagger struct {
	*tagging.BaseTagger
	TagManager *ArabicTagManager
}

func NewArabicTagger(wt tagging.WordTagger) *ArabicTagger {
	return &ArabicTagger{
		BaseTagger: tagging.NewBaseTagger(wt, ArabicDictPath, "ar", false),
		TagManager: NewArabicTagManager(),
	}
}

func (t *ArabicTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		striped := tools.RemoveTashkeel(word)
		var readings []*languagetool.AnalyzedToken
		for _, tw := range t.TagWord(striped) {
			readings = append(readings, tagged(word, tw))
		}
		// also try surface as-is if strip changed nothing useful
		if len(readings) == 0 && striped != word {
			for _, tw := range t.TagWord(word) {
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

// IsStopWordReading reports if any reading is a particle (P…).
func (t *ArabicTagger) IsStopWordReading(readings []*languagetool.AnalyzedToken) bool {
	if t == nil || t.TagManager == nil {
		return false
	}
	for _, r := range readings {
		if r != nil && r.GetPOSTag() != nil && t.TagManager.IsStopWord(*r.GetPOSTag()) {
			return true
		}
	}
	return false
}
