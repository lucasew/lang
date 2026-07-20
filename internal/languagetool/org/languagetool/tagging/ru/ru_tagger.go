package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

const RussianDictPath = "/ru/russian.dict"

// RussianTagger ports org.languagetool.tagging.ru.RussianTagger.
type RussianTagger struct {
	*tagging.BaseTagger
}

func NewRussianTagger(wt tagging.WordTagger) *RussianTagger {
	// Java BaseTagger("/ru/russian.dict", Locale("ru")) — tagLowercaseWithUppercase true by default.
	return &RussianTagger{BaseTagger: tagging.NewBaseTagger(wt, RussianDictPath, "ru", true)}
}

func (t *RussianTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		// Java: accent strip + ʼ→ъ, then getAnalyzedTokens (BaseTagger case-merge).
		norm, mayYo := tagging.NormalizeRussianSurface(word)
		var readings []*languagetool.AnalyzedToken
		for _, tw := range t.TagWord(norm) {
			// Readings keep original surface form (Java asAnalyzedTokenListForTaggedWords(word, …)).
			readings = append(readings, tagged(word, tw))
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
		}
		atr := languagetool.NewAnalyzedTokenReadingsList(readings, pos)
		if tagging.RussianMayMissingYoConfirmed(norm, mayYo, t.GetWordTagger()) {
			// Java: atr.setChunkTags([MayMissingYO])
			// AnalyzedTokenReadings chunk tags — set if API exists.
			if atr != nil {
				atr.SetChunkTags([]string{"MayMissingYO"})
			}
		}
		out = append(out, atr)
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
