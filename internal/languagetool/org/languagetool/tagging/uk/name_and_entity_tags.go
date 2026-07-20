package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	// common Ukrainian surname suffixes (unused for invent; surnames need dict)
	reNameSuffix = regexp.MustCompile(`(?i).+(енко|єнко|ишин|ішин|ський|цький|івна|ович|евич|ів|їв)$`)
	// x-shaped: Т-подібний handled elsewhere
	reXShaped = regexp.MustCompile(`(?i)^[\p{L}]-подібн`)
	// Java getAdjustedAnalyzedTokens ALLCAPS filter: Pattern.compile("noun.*?:prop.*|noninfl.*")
	// Matcher.matches() — full POS string only (not substring Contains).
	allCapsPropPOSRE = regexp.MustCompile(`^(?:noun.*?:prop.*|noninfl.*)$`)
)

// NameSuffixPOS is intentionally empty: Java tags surnames via the dictionary
// (…енко etc. are in uk.dict), not by inventing prop:lname from surface suffixes.
// Kept for call-site compatibility / tests that document fail-closed behavior.
func NameSuffixPOS(token string) string {
	_ = token
	_ = reNameSuffix
	return ""
}

// NumberedEntityPOS is deprecated invent-free: use EntityReadings (official entities.txt).
// Returns first POS if any entity pattern matches, else "".
func NumberedEntityPOS(token string) string {
	rs := EntityReadings(token)
	if len(rs) == 0 || rs[0].GetPOSTag() == nil {
		return ""
	}
	return *rs[0].GetPOSTag()
}

// isAllUppercaseUk ports LemmaHelper.isAllUppercaseUk (allows - – ' combining marks).
func isAllUppercaseUk(word string) bool {
	if word == "" {
		return false
	}
	for _, ch := range word {
		if ch == '-' || ch == '\u2013' || ch == '\'' || ch == '\u0301' || ch == '\u00AD' {
			continue
		}
		if !unicode.IsUpper(ch) {
			return false
		}
	}
	return true
}

// capitalizeProperName ports LemmaHelper.capitalizeProperName (title-case after '-').
func capitalizeProperName(word string) string {
	rs := []rune(word)
	if len(rs) == 0 {
		return word
	}
	out := make([]rune, len(rs))
	prev := '-'
	for i, ch := range rs {
		if prev == '-' {
			out[i] = unicode.ToUpper(ch)
		} else {
			out[i] = unicode.ToLower(ch)
		}
		if ch == '\u2013' {
			prev = '-'
		} else {
			prev = ch
		}
	}
	return string(out)
}

// AllCapsPropReadings ports UkrainianTagger ALLCAPS → capitalizeProperName + dict re-tag
// for noun.*:prop|noninfl. Fail closed without dictionary (no soft invent prop POS).
func AllCapsPropReadings(token string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || len([]rune(token)) <= 2 || !isAllUppercaseUk(token) {
		return nil
	}
	// skip pure latin romans (handled as number:latin)
	if reLatinNum.MatchString(token) {
		return nil
	}
	adjusted := capitalizeProperName(token)
	if adjusted == "" || adjusted == token {
		return nil
	}
	tws := tagWord(adjusted)
	if len(tws) == 0 {
		// also try lower lemma form in inject maps
		low := strings.ToLower(adjusted)
		if low != adjusted {
			tws = tagWord(low)
		}
	}
	if len(tws) == 0 {
		return nil
	}
	var out []*languagetool.AnalyzedToken
	for _, tw := range tws {
		pos := tw.PosTag
		if pos == "" {
			continue
		}
		// Java Pattern.compile("noun.*?:prop.*|noninfl.*") Matcher.matches()
		loc := allCapsPropPOSRE.FindStringIndex(pos)
		if loc == nil || loc[0] != 0 || loc[1] != len(pos) {
			continue
		}
		lemma := tw.Lemma
		if lemma == "" {
			lemma = adjusted
		}
		p, l := pos, lemma
		out = append(out, languagetool.NewAnalyzedToken(token, &p, &l))
	}
	return out
}

func utf8Decode(s string) (rune, int) {
	for _, r := range s {
		return r, 1
	}
	return 0, 0
}
