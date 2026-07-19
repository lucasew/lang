package uk

import (
	"regexp"
	"strings"
	"unicode"
)

// Dynamic numeric compounds: 100-й, 50-х, 100-річному, 100-відсотково
// (Java LetterEndingForNumericHelper / digit compounds — ending paradigms only).
var (
	// number(+optional %)+hyphen+short letter ending (ordinal-like)
	reNumOrd = regexp.MustCompile(`^(\d+(?:-\d+)?)(%?)[-–]([а-яіїєґА-ЯІЇЄҐ]+)$`)
	// number + hyphen + word
	reNumWord = regexp.MustCompile(`^(\d+(?:-\d+)?)[-–]([\p{L}].*)$`)
)

// ordinal short endings → adj:numr cases
var numrOrdExtra = map[string][]string{
	"х":   {":p:v_mis", ":p:v_rod", ":p:v_zna:ranim"},
	"ту":  {":f:v_zna"},
	"ій":  {":f:v_dav", ":f:v_mis", ":m:v_naz", ":m:v_zna:rinanim"},
	"ми":  {":p:v_oru"},
	"ні":  {":p:v_naz", ":p:v_zna:rinanim"},
	"ма":  {":f:v_naz"},
	"ти":  {":p:v_dav", ":p:v_mis", ":p:v_rod"},
	"ці":  {":f:v_dav", ":f:v_mis"},
	"ві":  {":p:v_naz", ":p:v_zna:rinanim"},
	"ому": {":m:v_dav", ":m:v_mis", ":n:v_dav", ":n:v_mis"},
	"ого": {":m:v_rod", ":m:v_zna:ranim", ":n:v_rod"},
	"им":  {":m:v_oru", ":n:v_oru", ":p:v_dav"},
	"ою":  {":f:v_oru"},
	"а":   {":f:v_naz"},
	"у":   {":f:v_zna"},
	"й":   {":m:v_naz", ":m:v_zna:rinanim", ":f:v_dav", ":f:v_mis"},
}

// DynamicNumericReadings tags digit-hyphen ordinal/adj endings (100-й, 100-річному).
// Does not invent bare noun POS for arbitrary right halves (10-хвилинка needs dict).
func DynamicNumericReadings(token string) []struct{ Lemma, POS string } {
	t := strings.ReplaceAll(token, "–", "-")
	t = strings.ReplaceAll(t, "—", "-")

	if m := reNumOrd.FindStringSubmatch(t); m != nil {
		num, pct, end := m[1], m[2], strings.ToLower(m[3])
		// short endings only for pure ordinals
		if len([]rune(end)) <= 3 {
			cases := numrOrdExtra[end]
			if len(cases) == 0 {
				cases = CasesForNumericEnding(num, end)
			}
			if len(cases) > 0 {
				lemma := num + pct + "-й"
				var out []struct{ Lemma, POS string }
				for _, c := range cases {
					out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: "adj" + c + ":numr"})
				}
				return out
			}
		}
	}

	if m := reNumWord.FindStringSubmatch(t); m != nil {
		num, right := m[1], m[2]
		if adj := numericAdjReadings(num, right); len(adj) > 0 {
			return adj
		}
		// adv: ends with -о (Java digit-adv compounds like 100-відсотково)
		low := strings.ToLower(right)
		if strings.HasSuffix(low, "о") && len([]rune(right)) > 3 {
			lemma := num + "-" + low
			return []struct{ Lemma, POS string }{{Lemma: lemma, POS: "adv"}}
		}
		// no invent noun:inanim for arbitrary "N-word" (fail closed without dict)
	}
	return nil
}

func numericAdjReadings(num, right string) []struct{ Lemma, POS string } {
	low := strings.ToLower(right)
	// Longest-first multi-letter adj endings only (ому/ого/…/ий).
	// Single-letter endings (а/у/е/і) invent too many noun surfaces as adj — fail closed.
	// Short ordinals (й/х/ту) use reNumOrd + numrOrdExtra instead.
	ends := []string{"ому", "ого", "ими", "их", "им", "ім", "ої", "ою", "ій", "ий"}
	for _, end := range ends {
		if !strings.HasSuffix(low, end) {
			continue
		}
		cases := adjEndingPOS[end]
		if len(cases) == 0 {
			continue
		}
		rs := []rune(right)
		ers := []rune(end)
		if len(rs) <= len(ers)+1 {
			continue
		}
		// require stem to look like adj base (ends with typical adj consonants)
		stemRunes := rs[:len(rs)-len(ers)]
		last := stemRunes[len(stemRunes)-1]
		if !strings.ContainsRune("нквлтсчгжшщрмдпбфц", unicode.ToLower(last)) {
			continue
		}
		lemma := num + "-" + strings.ToLower(string(stemRunes)) + "ий"
		var out []struct{ Lemma, POS string }
		for _, c := range cases {
			out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: "adj" + c})
		}
		return out
	}
	return nil
}

// MissingApostropheCandidates inserts ' before ї/є/ю/я after a consonant.
func MissingApostropheCandidates(token string) []string {
	if strings.Contains(token, "'") || strings.Contains(token, "’") {
		return nil
	}
	rs := []rune(token)
	var out []string
	needApo := map[rune]bool{'ї': true, 'є': true, 'ю': true, 'я': true, 'Ї': true, 'Є': true, 'Ю': true, 'Я': true}
	consonants := "бвгґджзклмнпрстфхцчшщБВГҐДЖЗКЛМНПРСТФХЦЧШЩ"
	for i := 1; i < len(rs); i++ {
		if !needApo[rs[i]] {
			continue
		}
		if !strings.ContainsRune(consonants, rs[i-1]) {
			continue
		}
		out = append(out, string(rs[:i])+"'"+string(rs[i:]))
	}
	return out
}
