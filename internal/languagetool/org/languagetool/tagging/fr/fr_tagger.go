package fr

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const FrenchDictPath = "/fr/french.dict"

var (
	frVerb             = regexp.MustCompile(`^V .+$`)
	frPrefixesForVerbs = regexp.MustCompile(`(?i)^(auto|auto-|re-|sur-)([^-].*[aeiouêàéèíòóïü].+[aeiouêàéèíòóïü].*)$`)
	frNounAdj          = regexp.MustCompile(`^[NJ] .+|V ppa.*$`)
	frPrefixesNounAdj  = regexp.MustCompile(`(?i)^(post-|sur-|mini-|méga-|demi-|péri-|anti-|géo-|nord-|sud-|néo-|méga-|ultra-|pro-|inter-|micro-|macro-|sous-|haut-|auto-|ré-|pré-|super-|vice-|hyper-|proto-|grand-|pseudo-)(.+)$`)
	frPrefixesForNounAdj = regexp.MustCompile(`(?i)^(mini|méga)([^-].*[aeiouêàéèíòóïü].+[aeiouêàéèíòóïü].*)$`)
)

// French ambiguous hyphen clitics that must not use hyphenated-title-case merge.
var frAmbiguousTokens = map[string]struct{}{
	"-Le": {}, "-Les": {}, "-La": {}, "-Elle": {}, "-Elles": {}, "-On": {},
	"-Tu": {}, "-Vous": {}, "-Il": {}, "-Ils": {}, "-Ce": {},
}

// FrenchTagger ports org.languagetool.tagging.fr.FrenchTagger.
type FrenchTagger struct {
	*tagging.BaseTagger
}

func NewFrenchTagger(wt tagging.WordTagger) *FrenchTagger {
	// Java: super("/fr/french.dict", Locale.FRENCH, false)
	return &FrenchTagger{BaseTagger: tagging.NewBaseTagger(wt, FrenchDictPath, "fr", false)}
}

func (t *FrenchTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		// Java: apostrophe hack + chunk tags (not setTypographicApostrophe).
		containsTypewriterApostrophe := false
		containsTypographicApostrophe := false
		w := word
		if len(w) > 1 {
			if strings.Contains(w, "'") {
				containsTypewriterApostrophe = true
			}
			if strings.Contains(w, "’") {
				containsTypographicApostrophe = true
				w = strings.ReplaceAll(w, "’", "'")
			}
		}
		readings := t.tagWord(w, w)
		if len(readings) == 0 && strings.Contains(strings.ToLower(w), "oe") {
			// Java: word.replace("oe","œ").replace("OE","Œ") then tagWord(..., word)
			alt := strings.ReplaceAll(strings.ReplaceAll(w, "oe", "œ"), "OE", "Œ")
			readings = t.tagWord(alt, w)
		}
		if len(readings) == 0 && tools.IsEmoji(word) {
			p, l := "_emoji_", "_emoji_"
			// Java uses (possibly replaced) word surface after apostrophe replace.
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(w, &p, &l)}
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(w, nil, nil)}
		}
		atr := languagetool.NewAnalyzedTokenReadingsList(readings, pos)
		// Java: setChunkTags replaces list; typographic overwrites typewriter when both.
		if containsTypewriterApostrophe {
			atr.SetChunkTags([]string{"containsTypewriterApostrophe"})
		}
		if containsTypographicApostrophe {
			atr.SetChunkTags([]string{"containsTypographicApostrophe"})
		}
		out = append(out, atr)
		pos += tagging.UTF16Len(w)
	}
	return out
}

// tagWord ports FrenchTagger.tagWord (exact + capitalized/all-upper/hyphen-title + first-upper + prefixes).
func (t *FrenchTagger) tagWord(word, originalWord string) []*languagetool.AnalyzedToken {
	if t == nil {
		return nil
	}
	var l []*languagetool.AnalyzedToken
	lowerWord := strings.ToLower(word)
	isStartUpper := tools.IsCapitalizedWord(word)
	isAllUpper := tools.IsAllUppercase(word)
	_, ambig := frAmbiguousTokens[originalWord]
	isHyphenatedTitleCase := !ambig && strings.Contains(originalWord, "-") &&
		originalWord == tools.ConvertToTitleCaseIteratingChars(lowerWord)

	// normal case
	for _, tw := range t.TagWordExact(word) {
		l = append(l, tagged(originalWord, tw))
	}
	// tag non-lowercase (alluppercase, startuppercase, hyphenated title case)
	if isAllUpper || isStartUpper || isHyphenatedTitleCase {
		for _, tw := range t.TagWordExact(lowerWord) {
			l = append(l, tagged(originalWord, tw))
		}
	}
	// all-uppercase proper nouns (ex. FRANCE)
	if len(l) == 0 && isAllUpper {
		firstUpper := tools.ConvertToTitleCaseIteratingChars(lowerWord)
		for _, tw := range t.TagWordExact(firstUpper) {
			l = append(l, tagged(originalWord, tw))
		}
	}
	// additional tagging with prefixes
	if len(l) == 0 {
		l = append(l, t.additionalTags(word)...)
	}
	return l
}

func (t *FrenchTagger) additionalTags(word string) []*languagetool.AnalyzedToken {
	if t == nil {
		return nil
	}
	// verb prefixes: auto|auto-|re-|sur-
	if m := frPrefixesForVerbs.FindStringSubmatch(word); m != nil {
		possibleVerb := strings.ToLower(m[2])
		var out []*languagetool.AnalyzedToken
		for _, tw := range t.TagWordExact(possibleVerb) {
			if tw.PosTag != "" && frVerb.MatchString(tw.PosTag) {
				p := tw.PosTag
				lemma := strings.ToLower(m[1]) + tw.Lemma
				out = append(out, languagetool.NewAnalyzedToken(word, &p, &lemma))
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	// mini|méga + noun/adj (no hyphen)
	if m := frPrefixesForNounAdj.FindStringSubmatch(word); m != nil {
		possibleNoun := strings.ToLower(m[2])
		var out []*languagetool.AnalyzedToken
		for _, tw := range t.TagWordExact(possibleNoun) {
			if tw.PosTag != "" && frNounAdj.MatchString(tw.PosTag) {
				p := tw.PosTag
				lemma := strings.ToLower(m[1]) + tw.Lemma
				out = append(out, languagetool.NewAnalyzedToken(word, &p, &lemma))
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	// hyphenated noun/adj prefixes
	if m := frPrefixesNounAdj.FindStringSubmatch(word); m != nil {
		possibleNoun := strings.ToLower(m[2])
		var out []*languagetool.AnalyzedToken
		for _, tw := range t.TagWordExact(possibleNoun) {
			if tw.PosTag != "" && frNounAdj.MatchString(tw.PosTag) {
				p := tw.PosTag
				lemma := strings.ToLower(m[1]) + tw.Lemma
				out = append(out, languagetool.NewAnalyzedToken(word, &p, &lemma))
			}
		}
		// with lower case (Java re-tags possibleNoun.toLowerCase — already lower)
		if len(out) == 0 {
			for _, tw := range t.TagWordExact(strings.ToLower(possibleNoun)) {
				if tw.PosTag != "" && frNounAdj.MatchString(tw.PosTag) {
					p := tw.PosTag
					lemma := strings.ToLower(m[1]) + tw.Lemma
					out = append(out, languagetool.NewAnalyzedToken(word, &p, &lemma))
				}
			}
		}
		return out
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
