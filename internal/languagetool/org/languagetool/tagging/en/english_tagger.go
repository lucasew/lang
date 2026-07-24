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
		// Java EnglishTagger: mutates local word to typewriter apostrophe for lookup +
		// AnalyzedToken surface; flags ATR with setTypographicApostrophe.
		containsTypographicApostrophe := false
		if len(word) > 1 && strings.Contains(word, "’") {
			containsTypographicApostrophe = true
			word = strings.ReplaceAll(word, "’", "'")
		}
		lower := strings.ToLower(word)
		isLower := word == lower
		isMixed := tools.IsMixedCase(word)
		isAllUpper := tools.IsAllUppercase(word)

		// Java EnglishTagger.tag uses getWordTagger().tag (exact), not BaseTagger.getAnalyzedTokens.
		// TagWord would re-merge case variants and duplicate title-case readings.
		var readings []*languagetool.AnalyzedToken
		for _, tw := range t.TagWordExact(word) {
			readings = append(readings, taggedToToken(word, tw))
		}
		if !isLower && !isMixed {
			for _, tw := range t.TagWordExact(lower) {
				readings = append(readings, taggedToToken(word, tw))
			}
		}
		if len(readings) == 0 && isAllUpper {
			firstUpper := tools.UppercaseFirstChar(lower)
			for _, tw := range t.TagWordExact(firstUpper) {
				readings = append(readings, taggedToToken(word, tw))
			}
		}
		// Java: walkin'/doin' → walking/doing style (endsWith "in'")
		if len(readings) == 0 && strings.HasSuffix(lower, "in'") {
			corrected := word
			if isAllUpper {
				corrected = word[:len(word)-1] + "G"
			} else {
				corrected = word[:len(word)-1] + "g"
			}
			for _, tw := range t.TagWordExact(corrected) {
				readings = append(readings, taggedToToken(word, tw))
			}
			if !isLower && !isMixed {
				for _, tw := range t.TagWordExact(strings.ToLower(corrected)) {
					readings = append(readings, taggedToToken(word, tw))
				}
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
		// Java: pos += word.length() after apostrophe rewrite (same UTF-16 length for ’→').
		pos += tagging.UTF16Len(word)
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
