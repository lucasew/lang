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

// Tag ports RussianTagger.tag: accent/hard-sign strip, BaseTagger.getAnalyzedTokens,
// MayMissingYO chunk tag when the е→ё dictionary probe hits.
func (t *RussianTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		// Java: compute mayMissingYo on original surface, then mutate word in place
		// (acute/grave strip, ʼ→ъ) before getAnalyzedTokens / pos increment.
		norm, mayYo := tagging.NormalizeRussianSurface(word)
		word = norm
		var readings []*languagetool.AnalyzedToken
		// Java getAnalyzedTokens(word) — BaseTagger case-merge via TagWord.
		for _, tw := range t.TagWord(word) {
			readings = append(readings, tagged(word, tw))
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
		}
		atr := languagetool.NewAnalyzedTokenReadingsList(readings, pos)
		// Java: getWordTagger().tag(wordLc with е→ё); empty → clear flag.
		if tagging.RussianMayMissingYoConfirmed(word, mayYo, t.GetWordTagger()) {
			// Java: atr.setChunkTags([new ChunkTag("MayMissingYO")])
			atr.SetChunkTags([]string{"MayMissingYO"})
		}
		out = append(out, atr)
		// Java: pos += word.length() after mutation (UTF-16 code units).
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
