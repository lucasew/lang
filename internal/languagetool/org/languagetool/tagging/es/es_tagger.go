package es

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Java SpanishTagger uses /es/es-ES.dict (not spanish.dict).
const SpanishDictPath = "/es/es-ES.dict"

var (
	esAdjPartFS          = regexp.MustCompile(`^VMP00SF|A[QO].[FC]S.$`)
	esVerb               = regexp.MustCompile(`^V.+`)
	esPrefixesForVerbs   = regexp.MustCompile(`(?i)^(auto)([^r].{3,})$`)
	esPrefixesForVerbs2  = regexp.MustCompile(`(?i)^(autor)(r.{3,})$`)
	esPrefixesForAdj     = regexp.MustCompile(`(?i)^(.+)-(.+)$`)
	esAdj                = regexp.MustCompile(`^AQ.+`)
	esAdjMS              = regexp.MustCompile(`^AQ.MS.|AQ.CS.|AQ.MN.$`)
	esNoPrefixesForAdj   = regexp.MustCompile(`(?i)^(anti|pre|ex|pro|afro|ultra|super|súper)$`)
	esPrefixesForAdjs    = regexp.MustCompile(`(?i)^(super)(.*[aeiouáéèíòóïü].+[aeiouáéèíòóïü].*)$`)
	esAdjVP              = regexp.MustCompile(`^AQ.*|V.P.*$`)
)

// SpanishTagger ports org.languagetool.tagging.es.SpanishTagger.
type SpanishTagger struct {
	*tagging.BaseTagger
}

func NewSpanishTagger(wt tagging.WordTagger) *SpanishTagger {
	// Java BaseTagger default tagLowercaseWithUppercase=true; Tag reimplements case.
	return &SpanishTagger{BaseTagger: tagging.NewBaseTagger(wt, SpanishDictPath, "es", true)}
}

func (t *SpanishTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for i, word := range sentenceTokens {
		prev, next := "", ""
		if i > 0 {
			prev = sentenceTokens[i-1]
		}
		if i+1 < len(sentenceTokens) {
			next = sentenceTokens[i+1]
		}
		// Java: replace ’ when length>1 OR prev is l/d OR next is s
		containsTypographic := false
		w := word
		if len(w) > 1 || strings.EqualFold(prev, "l") || strings.EqualFold(prev, "d") ||
			strings.EqualFold(next, "s") {
			if strings.Contains(w, "’") {
				containsTypographic = true
				w = strings.ReplaceAll(w, "’", "'")
			}
		}
		lower := strings.ToLower(w)
		isLower := w == lower
		isMixed := tools.IsMixedCase(w)
		isAllUpper := tools.IsAllUppercase(w)

		var readings []*languagetool.AnalyzedToken
		// exact getWordTagger().tag
		for _, tw := range t.TagWordExact(w) {
			readings = append(readings, tagged(word, tw))
		}
		if !isLower && !isMixed {
			for _, tw := range t.TagWordExact(lower) {
				readings = append(readings, tagged(word, tw))
			}
		}
		// all-uppercase proper nouns (FRANCIA)
		if isAllUpper {
			firstUpper := tools.UppercaseFirstChar(lower)
			for _, tw := range t.TagWordExact(firstUpper) {
				readings = append(readings, tagged(word, tw))
			}
		}
		if len(readings) == 0 && !isMixed {
			for _, at := range t.additionalTags(w) {
				readings = append(readings, at)
			}
		}
		if len(readings) == 0 && tools.IsEmoji(word) {
			p, l := "_emoji_", "_emoji_"
			readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
		}
		atr := languagetool.NewAnalyzedTokenReadingsList(readings, pos)
		if containsTypographic {
			atr.SetTypographicApostrophe(true)
		}
		out = append(out, atr)
		pos += tagging.UTF16Len(word)
	}
	return out
}

func (t *SpanishTagger) additionalTags(word string) []*languagetool.AnalyzedToken {
	if t == nil {
		return nil
	}
	lower := strings.ToLower(word)
	// -mente adverbs
	if strings.HasSuffix(lower, "mente") {
		possibleAdj := strings.TrimSuffix(lower, "mente")
		for _, tw := range t.TagWordExact(possibleAdj) {
			if tw.PosTag != "" && esAdjPartFS.MatchString(tw.PosTag) {
				p, lemma := "RG", lower
				return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &lemma)}
			}
		}
	}
	// auto + verb (not autor…)
	if m := esPrefixesForVerbs.FindStringSubmatch(word); m != nil {
		possibleVerb := strings.ToLower(m[2])
		var out []*languagetool.AnalyzedToken
		for _, tw := range t.TagWordExact(possibleVerb) {
			if tw.PosTag != "" && esVerb.MatchString(tw.PosTag) {
				p := tw.PosTag
				lemma := strings.ToLower(m[1]) + tw.Lemma
				out = append(out, languagetool.NewAnalyzedToken(word, &p, &lemma))
			}
		}
		return out
	}
	// autor + r…
	if m := esPrefixesForVerbs2.FindStringSubmatch(word); m != nil {
		possibleVerb := strings.ToLower(m[2])
		var out []*languagetool.AnalyzedToken
		for _, tw := range t.TagWordExact(possibleVerb) {
			if tw.PosTag != "" && esVerb.MatchString(tw.PosTag) {
				p := tw.PosTag
				lemma := strings.ToLower(m[1]) + tw.Lemma
				out = append(out, languagetool.NewAnalyzedToken(word, &p, &lemma))
			}
		}
		return out
	}
	// super + adj/participle (two syllables)
	if m := esPrefixesForAdjs.FindStringSubmatch(word); m != nil {
		possible := strings.ToLower(m[2])
		var out []*languagetool.AnalyzedToken
		for _, tw := range t.TagWordExact(possible) {
			if tw.PosTag != "" && esAdjVP.MatchString(tw.PosTag) {
				p := tw.PosTag
				lemma := strings.ToLower(m[1]) + tw.Lemma
				out = append(out, languagetool.NewAnalyzedToken(word, &p, &lemma))
			}
		}
		return out
	}
	// adj-adj compounds
	if m := esPrefixesForAdj.FindStringSubmatch(word); m != nil {
		possibleAdjPrefix := strings.ToLower(m[1])
		if !esNoPrefixesForAdj.MatchString(possibleAdjPrefix) {
			possibleAdj := strings.ToLower(m[2])
			prefixMatches := false
			for _, tw := range t.TagWordExact(possibleAdjPrefix) {
				if tw.PosTag != "" && esAdjMS.MatchString(tw.PosTag) {
					prefixMatches = true
					break
				}
			}
			var newPostag, newLemma string
			adjMatches := false
			for _, tw := range t.TagWordExact(possibleAdj) {
				if tw.PosTag != "" && esAdj.MatchString(tw.PosTag) {
					adjMatches = true
					newPostag = tw.PosTag
					newLemma = possibleAdjPrefix + "-" + tw.Lemma
					break
				}
			}
			if adjMatches && prefixMatches {
				return []*languagetool.AnalyzedToken{
					languagetool.NewAnalyzedToken(word, &newPostag, &newLemma),
				}
			}
		}
	}
	return nil
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
