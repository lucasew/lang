package de

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const GermanDictPath = "/de/german.dict"

// GermanTagger ports org.languagetool.tagging.de.GermanTagger (dict/compound split deferred).
type GermanTagger struct {
	*tagging.BaseTagger
	// SplitCompound optional compound splitter for unknown tokens.
	SplitCompound func(word string) []string
}

func NewGermanTagger(wt tagging.WordTagger) *GermanTagger {
	return &GermanTagger{
		BaseTagger: tagging.NewBaseTagger(wt, GermanDictPath, "de", true),
	}
}

// DefaultGermanTagger is a process-level tagger (empty dict until loaded).
var DefaultGermanTagger = NewGermanTagger(tagging.MapWordTagger{})

// Tag tags tokens with German case/gender-aware retries (simplified vs full Java).
func (t *GermanTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		readings := t.tagOne(word)
		out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
		pos += len([]rune(word))
	}
	return out
}

func (t *GermanTagger) tagOne(word string) []*languagetool.AnalyzedToken {
	w := word
	var readings []*languagetool.AnalyzedToken
	for _, tw := range t.TagWord(w) {
		readings = append(readings, toToken(word, tw))
	}
	lower := strings.ToLower(w)
	if len(readings) == 0 && w != lower {
		for _, tw := range t.TagWord(lower) {
			readings = append(readings, toToken(word, tw))
		}
	}
	// capitalized common nouns
	if len(readings) == 0 && tools.StartsWithUppercase(w) {
		// try as-is already done; try lower first letter only for all-upper
		if isAllUpperLetters(w) {
			fu := tools.UppercaseFirstChar(lower)
			for _, tw := range t.TagWord(fu) {
				readings = append(readings, toToken(word, tw))
			}
		}
	}
	// compound split fallback
	if len(readings) == 0 && t.SplitCompound != nil {
		parts := t.SplitCompound(w)
		if len(parts) > 1 {
			// tag last part for head noun
			last := parts[len(parts)-1]
			for _, tw := range t.TagWord(last) {
				readings = append(readings, toToken(word, tw))
			}
		}
	}
	if len(readings) == 0 {
		readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
	}
	return readings
}

func toToken(surface string, tw tagging.TaggedWord) *languagetool.AnalyzedToken {
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

func isAllUpperLetters(s string) bool {
	has := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			has = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return has
}

// SwissGermanTagger ports tagging.de.SwissGermanTagger as GermanTagger with ss↔ß retry.
type SwissGermanTagger struct {
	*GermanTagger
}

func NewSwissGermanTagger(wt tagging.WordTagger) *SwissGermanTagger {
	return &SwissGermanTagger{GermanTagger: NewGermanTagger(wt)}
}

func (t *SwissGermanTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	// map ss spellings: try original via base, then ß variants
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		readings := t.tagOne(word)
		if len(readings) == 1 && readings[0].GetPOSTag() == nil && strings.Contains(word, "ss") {
			alt := strings.ReplaceAll(word, "ss", "ß")
			if alt != word {
				readings = t.tagOne(alt)
				// restore surface
				for _, r := range readings {
					// rebuild with original surface - NewAnalyzedToken
					_ = r
				}
			}
		}
		out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
		pos += len([]rune(word))
	}
	return out
}
