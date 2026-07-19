package uk

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// leftOAdjInvalid ports CompoundTagger.LEFT_O_ADJ_INVALID — directional / degree
// prefixes that normally write solid; hyphen form is :bad unless both parts are
// capitalized and the full lower compound is already an adj in the dict.
var leftOAdjInvalid = map[string]struct{}{
	"багато": {}, "мало": {}, "високо": {}, "низько": {}, "старо": {}, "важко": {},
	"зовнішньо": {}, "внутрішньо": {}, "ново": {}, "середньо": {},
	"південно": {}, "північно": {}, "західно": {}, "східно": {}, "центрально": {},
	"ранньо": {}, "пізньо": {},
}

// Directional left stems commonly produced by Java oAdjMatch for geo/direction compounds.
var directionalLeft = map[string]struct{}{
	"південно": {}, "північно": {}, "східно": {}, "західно": {},
	"центрально": {}, "ранньо": {}, "пізньо": {},
}

// DynamicDirectionalAdjReadings ports CompoundTagger.oAdjMatch for directional lefts.
// Requires tagWord hits on the right part as adj (Java wordTagger); fail-closed without dict.
// Does not invent case endings from surface alone.
func DynamicDirectionalAdjReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	if strings.Count(token, "-") != 1 {
		return nil
	}
	dash := strings.LastIndex(token, "-")
	if dash <= 0 || dash == len(token)-1 {
		return nil
	}
	leftWord := token[:dash]
	rightWord := token[dash+1:]
	if utf8.RuneCountInString(leftWord) < 3 {
		return nil
	}
	leftLow := strings.ToLower(leftWord)
	// Gate shape: known directional left, or ends with о/е like Java O_ADJ_PATTERN.
	if _, ok := directionalLeft[leftLow]; !ok {
		if !strings.HasSuffix(leftLow, "о") && !strings.HasSuffix(leftLow, "е") {
			return nil
		}
	}

	tws := tagWord(rightWord)
	if len(tws) == 0 {
		low := strings.ToLower(rightWord)
		if low != rightWord {
			tws = tagWord(low)
		}
	}
	if len(tws) == 0 {
		return nil
	}

	// Java analyzeAllCapitamizedAdj: both parts capitalized → if full lower is adj, skip :bad.
	skipBadForInvalid := false
	if isCapitalizedWord(leftWord) && isCapitalizedWord(rightWord) {
		for _, tw := range tagWord(strings.ToLower(token)) {
			if strings.HasPrefix(tw.PosTag, "adj") {
				skipBadForInvalid = true
				break
			}
		}
	}

	_, leftInvalid := leftOAdjInvalid[leftLow]
	extraBad := leftInvalid && !skipBadForInvalid

	var out []struct{ Lemma, POS string }
	seen := map[string]struct{}{}
	for _, tw := range tws {
		pos := tw.PosTag
		if pos == "" || !strings.HasPrefix(pos, "adj") {
			continue
		}
		// Java: drop :comp. from right before combining
		if i := strings.Index(pos, ":comp"); i >= 0 {
			end := i + len(":comp")
			for end < len(pos) && pos[end] != ':' {
				end++
			}
			pos = pos[:i] + pos[end:]
		}
		if extraBad {
			pos = strings.ReplaceAll(pos, ":arch", "")
			if !strings.Contains(pos, ":bad") {
				pos = pos + ":bad"
			}
		}
		rightLemma := tw.Lemma
		if rightLemma == "" {
			rightLemma = strings.ToLower(rightWord)
		}
		lemma := leftLow + "-" + rightLemma
		key := lemma + "|" + pos
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: pos})
	}
	return out
}

func isCapitalizedWord(s string) bool {
	if s == "" {
		return false
	}
	rs := []rune(s)
	if !unicode.IsUpper(rs[0]) {
		return false
	}
	for _, r := range rs[1:] {
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

// MissingHyphenCandidates returns alternate surfaces to try when word is untagged
// (e.g. insert hyphen after known prefix, or before -небудь).
func MissingHyphenCandidates(token string) []string {
	if strings.Contains(token, "-") {
		return nil
	}
	lower := strings.ToLower(token)
	var out []string
	for _, prefix := range []string{"напів", "пів", "анти", "псевдо", "міні", "віце", "екс"} {
		if !strings.HasPrefix(lower, prefix) || len(lower) <= len(prefix)+1 {
			continue
		}
		rs := []rune(token)
		pr := []rune(prefix)
		if len(rs) <= len(pr) {
			continue
		}
		next := rs[len(pr)]
		if !unicode.IsLetter(next) {
			continue
		}
		cand := string(rs[:len(pr)]) + "-" + string(rs[len(pr):])
		out = append(out, cand)
	}
	if strings.HasSuffix(lower, "небудь") && len([]rune(lower)) > len([]rune("небудь"))+2 {
		rs := []rune(token)
		suf := []rune("небудь")
		stem := string(rs[:len(rs)-len(suf)])
		out = append(out, stem+"-небудь")
	}
	return out
}

// CompoundNumrReadings tags forms like "2-х", "3-ом" soft.
var reCompoundNumr = regexp.MustCompile(`^(\d+)([-–])?(х|ом|им|и|а|е|го|му)?$`)

func CompoundNumrPOS(token string) string {
	if reCompoundNumr.MatchString(token) && strings.ContainsAny(token, "0123456789") {
		hasLetter := false
		for _, r := range token {
			if unicode.IsLetter(r) {
				hasLetter = true
				break
			}
		}
		if hasLetter {
			return "numr"
		}
	}
	return ""
}
