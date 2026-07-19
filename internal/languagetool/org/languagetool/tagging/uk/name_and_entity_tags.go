package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	// common Ukrainian surname suffixes
	reNameSuffix = regexp.MustCompile(`(?i).+(енко|єнко|ишин|ішин|ський|цький|івна|ович|евич|ів|їв)$`)
	// numbered entities: Т-80, Ан-225, Як-42
	reNumberedEntity = regexp.MustCompile(`^[\p{L}]{1,4}-?\d{1,4}[А-Яа-яA-Za-z]?$`)
	// x-shaped: Т-подібний handled elsewhere; simple letter-dash-letter
	reXShaped = regexp.MustCompile(`(?i)^[\p{L}]-подібн`)
)

// NameSuffixPOS returns a prop name POS for surname-like tokens, or "".
func NameSuffixPOS(token string) string {
	if !reNameSuffix.MatchString(token) {
		return ""
	}
	// require initial capital for prop names
	r, _ := utf8Decode(token)
	if !unicode.IsUpper(r) {
		return ""
	}
	return "noun:anim:m:v_naz:prop:lname"
}

// NumberedEntityPOS tags military/aircraft style designations.
func NumberedEntityPOS(token string) string {
	if reNumberedEntity.MatchString(token) && strings.ContainsAny(token, "0123456789") {
		// must have a letter part
		hasL := false
		for _, r := range token {
			if unicode.IsLetter(r) {
				hasL = true
				break
			}
		}
		if hasL {
			return "noun:inanim:m:v_naz:prop:unanim"
		}
	}
	return ""
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
		// Java Pattern: noun.*?:prop.*|noninfl.*
		ok := (strings.Contains(pos, "noun") && strings.Contains(pos, "prop")) ||
			strings.HasPrefix(pos, "noninfl")
		if !ok {
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
