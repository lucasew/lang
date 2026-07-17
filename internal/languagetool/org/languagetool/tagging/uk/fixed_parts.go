package uk

import (
	"strings"
	"unicode"
)

// Street / avenue right-hand fixed parts (Java CompoundTagger street suffixes).
var fixedRightParts = map[string]string{
	"авеню": "f",
	"стрит": "f",
	"стріт": "f",
	"сквер": "m",
	"плаза": "f",
}

var nvCasesF = []string{":f:v_dav:nv", ":f:v_mis:nv", ":f:v_naz:nv", ":f:v_oru:nv", ":f:v_rod:nv", ":f:v_zna:nv"}
var nvCasesP = []string{":p:v_dav:nv", ":p:v_mis:nv", ":p:v_naz:nv", ":p:v_oru:nv", ":p:v_rod:nv", ":p:v_zna:nv"}
var nvCasesM = []string{":m:v_dav:nv", ":m:v_mis:nv", ":m:v_naz:nv", ":m:v_oru:nv", ":m:v_rod:nv", ":m:v_zna:nv"}

// FixedPartReadings tags пів-X and Name-авеню/стрит style compounds without full dict.
func FixedPartReadings(token string) []struct{ Lemma, POS string } {
	t := strings.ReplaceAll(token, "–", "-")
	t = strings.ReplaceAll(t, "—", "-")
	rs := []rune(t)
	if len(rs) < 5 {
		return nil
	}

	// пів-України / пів-години
	if strings.HasPrefix(strings.ToLower(string(rs[:4])), "пів-") || (len(rs) >= 4 && strings.EqualFold(string(rs[:3]), "пів") && rs[3] == '-') {
		// split after first hyphen following пів
		idx := -1
		for i, r := range rs {
			if r == '-' {
				idx = i
				break
			}
		}
		if idx > 0 && idx+1 < len(rs) {
			right := string(rs[idx+1:])
			lemma := strings.ToLower(t)
			extra := ":bad"
			if rr := []rune(right); len(rr) > 0 && unicode.IsUpper(rr[0]) {
				extra = ":prop:geo:alt"
			}
			var out []struct{ Lemma, POS string }
			for _, c := range nvCasesP {
				out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: "noun:inanim" + c + extra})
			}
			return out
		}
	}

	// Name-авеню / Name-стрит
	if i := strings.LastIndex(t, "-"); i > 0 {
		right := strings.ToLower(t[i+1:])
		gender, ok := fixedRightParts[right]
		if !ok {
			return nil
		}
		cases := nvCasesF
		if gender == "m" {
			cases = nvCasesM
		}
		lemma := t
		var out []struct{ Lemma, POS string }
		for _, c := range cases {
			out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: "noun:inanim" + c + ":prop"})
		}
		return out
	}
	return nil
}
