package uk

import (
	"regexp"
	"strings"
	"unicode"
)

// Південно-Західній style directional compounds
var reDirectional = regexp.MustCompile(
	`(?i)^(південно|північно|східно|західно)-(західн|східн|північн|південн)(ий|ій|ого|ому|им|ім|а|ої|у|ою|е|і|их|ими)$`,
)

// DynamicDirectionalAdjReadings tags capitalized directional compounds.
func DynamicDirectionalAdjReadings(token string) []struct{ Lemma, POS string } {
	m := reDirectional.FindStringSubmatch(token)
	if m == nil {
		return nil
	}
	// lemma: full lower with -ий
	stem := strings.ToLower(m[1] + "-" + m[2])
	lemma := stem + "ий"
	end := strings.ToLower(m[3])
	// map ій as soft ending for m/f
	cases := adjEndingPOS[end]
	if end == "ій" {
		// can be f:v_dav/mis or m:v_naz bad forms — include both
		cases = append(cases, ":f:v_dav", ":f:v_mis", ":m:v_naz", ":m:v_zna:rinanim")
	}
	if len(cases) == 0 {
		cases = []string{":m:v_naz"}
	}
	var out []struct{ Lemma, POS string }
	seen := map[string]struct{}{}
	for _, c := range cases {
		pos := "adj" + c
		if _, ok := seen[pos]; ok {
			continue
		}
		seen[pos] = struct{}{}
		out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: pos})
	}
	return out
}

// MissingHyphenCandidates returns alternate surfaces to try when word is untagged
// (e.g. insert hyphen after known prefix).
func MissingHyphenCandidates(token string) []string {
	lower := strings.ToLower(token)
	var out []string
	for _, prefix := range []string{"напів", "пів", "анти", "псевдо", "міні", "віце", "екс"} {
		if !strings.HasPrefix(lower, prefix) || len(lower) <= len(prefix)+1 {
			continue
		}
		// already hyphenated?
		if strings.Contains(token, "-") {
			continue
		}
		// insert hyphen after prefix
		rs := []rune(token)
		pr := []rune(prefix)
		if len(rs) <= len(pr) {
			continue
		}
		// only if next char is uppercase (missing hyphen after prefix before proper) or letter
		next := rs[len(pr)]
		if !unicode.IsLetter(next) {
			continue
		}
		cand := string(rs[:len(pr)]) + "-" + string(rs[len(pr):])
		out = append(out, cand)
	}
	return out
}

// CompoundNumrReadings tags forms like "2-х", "3-ом" soft.
var reCompoundNumr = regexp.MustCompile(`^(\d+)([-–])?(х|ом|им|и|а|е|го|му)?$`)

func CompoundNumrPOS(token string) string {
	if reCompoundNumr.MatchString(token) && strings.ContainsAny(token, "0123456789") {
		// require letter ending for numr-like
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
