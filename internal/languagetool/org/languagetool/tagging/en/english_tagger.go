package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const EnglishDictPath = "/en/english.dict"

// EnglishTagger ports org.languagetool.tagging.en.EnglishTagger.
type EnglishTagger struct {
	*tagging.BaseTagger
	InternTags bool
}

// NewEnglishTagger builds an English tagger over the given WordTagger (dict deferred).
func NewEnglishTagger(wt tagging.WordTagger) *EnglishTagger {
	return &EnglishTagger{
		BaseTagger: tagging.NewBaseTagger(wt, EnglishDictPath, "en", false),
		InternTags: true,
	}
}

// DefaultEnglishTagger is the process singleton (Java INSTANCE).
var DefaultEnglishTagger = NewEnglishTagger(tagging.MapWordTagger{})

// Tag ports EnglishTagger.tag with typographic apostrophe normalisation and case retries.
func (t *EnglishTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		// Java EnglishTagger: typewriter apostrophe hack + setTypographicApostrophe.
		w := word
		containsTypographicApostrophe := false
		if len(w) > 1 && strings.Contains(w, "’") {
			containsTypographicApostrophe = true
			w = strings.ReplaceAll(w, "’", "'")
		}
		lower := strings.ToLower(w)
		isLower := w == lower
		isMixed := tools.IsMixedCase(w)
		isAllUpper := tools.IsAllUppercase(w)

		var readings []*languagetool.AnalyzedToken
		for _, tw := range t.TagWord(w) {
			readings = append(readings, taggedToToken(word, tw))
		}
		if !isLower && !isMixed {
			for _, tw := range t.TagWord(lower) {
				readings = append(readings, taggedToToken(word, tw))
			}
		}
		if len(readings) == 0 && isAllUpper {
			firstUpper := tools.UppercaseFirstChar(lower)
			for _, tw := range t.TagWord(firstUpper) {
				readings = append(readings, taggedToToken(word, tw))
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

func taggedToToken(surface string, tw tagging.TaggedWord) *languagetool.AnalyzedToken {
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
