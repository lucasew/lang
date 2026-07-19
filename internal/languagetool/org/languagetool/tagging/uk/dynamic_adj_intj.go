package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Dynamic adjective patterns: X-подібний / X-вмісний with case endings.
// (Java CompoundTagger / dynamic adj patterns — ending paradigms for those suffixes.)
var (
	rePodibny = regexp.MustCompile(`(?i)^(.+-подібн)(ий|ого|ому|им|ім|а|ої|ій|у|ою|е|і|их|ими)$`)
	reVmisny  = regexp.MustCompile(`(?i)^(.+-вмісн)(ий|ого|ому|им|ім|а|ої|ій|у|ою|е|і|их|ими)$`)
)

// ukVowels for elongated collapse (Java ([аеєиіїоуюя])\1{2,} — RE2 has no backrefs).
const ukVowels = "аеєиіїоуюяАЕЄИІЇОУЮЯ"

// collapseElongatedVowels replaces any vowel repeated 3+ times with a single copy.
func collapseElongatedVowels(token string) (adjusted string, changed bool) {
	if token == "" {
		return token, false
	}
	var b strings.Builder
	rs := []rune(token)
	for i := 0; i < len(rs); {
		r := rs[i]
		b.WriteRune(r)
		if strings.ContainsRune(ukVowels, r) {
			j := i + 1
			for j < len(rs) && rs[j] == r {
				j++
			}
			if j-i >= 3 {
				changed = true
				// already wrote one; skip the rest of the run
				i = j
				continue
			}
		}
		i++
	}
	return b.String(), changed
}

// ending → POS case list for adjectives (simplified soft paradigm)
var adjEndingPOS = map[string][]string{
	"ий":  {":m:v_naz", ":m:v_zna:rinanim"},
	"ого": {":m:v_rod", ":m:v_zna:ranim", ":n:v_rod"},
	"ому": {":m:v_dav", ":m:v_mis", ":n:v_dav", ":n:v_mis"},
	"им":  {":m:v_oru", ":n:v_oru", ":p:v_dav"},
	"ім":  {":m:v_mis", ":n:v_mis"},
	"а":   {":f:v_naz"},
	"ої":  {":f:v_rod"},
	"ій":  {":f:v_dav", ":f:v_mis"},
	"у":   {":f:v_zna"},
	"ою":  {":f:v_oru"},
	"е":   {":n:v_naz", ":n:v_zna"},
	"і":   {":p:v_naz", ":p:v_zna:rinanim"},
	"их":  {":p:v_rod", ":p:v_zna:ranim"},
	"ими": {":p:v_oru"},
}

// DynamicAdjReadings returns lemma|POS pairs for -подібний / -вмісний forms.
func DynamicAdjReadings(token string) []struct{ Lemma, POS string } {
	for _, re := range []*regexp.Regexp{rePodibny, reVmisny} {
		m := re.FindStringSubmatch(token)
		if m == nil {
			continue
		}
		stem, end := m[1], strings.ToLower(m[2])
		lemma := lowerFirst(stem + "ий")
		cases := adjEndingPOS[end]
		if len(cases) == 0 {
			continue
		}
		var out []struct{ Lemma, POS string }
		for _, c := range cases {
			out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: "adj" + c})
		}
		return out
	}
	return nil
}

// ElongatedAltReadings ports UkrainianTagger elongated-vowel collapse:
// when surface has a vowel repeated 3+ times, re-tag the collapsed form and mark :alt.
// Fail closed without dictionary hits (Java getAdjustedAnalyzedTokens; no invent intj list).
// Skips "ііі" (often Latin number III) like Java.
func ElongatedAltReadings(token string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || token == "" {
		return nil
	}
	if strings.EqualFold(token, "ііі") {
		return nil
	}
	adjusted, changed := collapseElongatedVowels(token)
	if !changed || adjusted == "" {
		return nil
	}
	// Java posTagRegex: (?!noun.*:prop|.*abbr).*
	tws := tagWord(adjusted)
	if len(tws) == 0 {
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
		// skip proper nouns and abbr (Java negative lookahead)
		if strings.Contains(pos, "noun") && strings.Contains(pos, "prop") {
			continue
		}
		if strings.Contains(pos, "abbr") {
			continue
		}
		if !strings.Contains(pos, ":alt") {
			pos = pos + ":alt"
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

func lowerFirst(s string) string {
	rs := []rune(s)
	if len(rs) == 0 {
		return s
	}
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}
