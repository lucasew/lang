package uk

import (
	"regexp"
	"strings"
	"unicode"
)

// Dynamic adjective patterns: X-подібний / X-вмісний with case endings.
var (
	rePodibny = regexp.MustCompile(`(?i)^(.+-подібн)(ий|ого|ому|им|ім|а|ої|ій|у|ою|е|і|их|ими)$`)
	reVmisny  = regexp.MustCompile(`(?i)^(.+-вмісн)(ий|ого|ому|им|ім|а|ої|ій|у|ою|е|і|их|ими)$`)
	// known interjection compounds
	knownIntjHyphen = map[string]string{
		"гей-но": "гей", "цить-но": "цить", "ану-бо": "ану",
		"а-а": "а-а", "га-га": "га-га", "фу-фу": "фу-фу",
		"гей-гей-гей": "гей-гей-гей", "ого-го-го-го": "ого-го-го-го",
	}
	// elongated interjection: same vowel/syllable run
	reElongIntj = regexp.MustCompile(`(?i)^(га|го|гей|а|о|у|е|и|фу)([аеєиіїоуюяь]{2,})$`)
)

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

// IntjReading returns intj POS for elongated or hyphenated interjections.
func IntjReading(token string) (lemma, pos string, ok bool) {
	low := strings.ToLower(token)
	if base, found := knownIntjHyphen[low]; found {
		return base, "intj", true
	}
	// repeated hyphen syllables: га-га-га
	if strings.Count(token, "-") >= 1 {
		parts := strings.Split(low, "-")
		allSame := true
		for i := 1; i < len(parts); i++ {
			if parts[i] != parts[0] {
				allSame = false
				break
			}
		}
		if allSame && len(parts[0]) >= 1 && len(parts[0]) <= 5 && isCyrillicWord(parts[0]) {
			return low, "intj", true
		}
	}
	// elongated: гаааа
	if m := reElongIntj.FindStringSubmatch(token); m != nil {
		return strings.ToLower(m[1]), "intj:alt", true
	}
	// pure vowel runs
	if isElongatedVowelRun(low) {
		r := []rune(low)[0]
		return string(r), "intj:alt", true
	}
	return "", "", false
}

func isCyrillicWord(s string) bool {
	for _, r := range s {
		if !unicode.Is(unicode.Cyrillic, r) {
			return false
		}
	}
	return s != ""
}

func isElongatedVowelRun(s string) bool {
	if len([]rune(s)) < 3 {
		return false
	}
	vowels := "аеєиіїоуюяь"
	first := true
	var base rune
	for _, r := range s {
		if !strings.ContainsRune(vowels, r) {
			return false
		}
		if first {
			base = r
			first = false
			continue
		}
		if r != base {
			return false
		}
	}
	return true
}

func lowerFirst(s string) string {
	rs := []rune(s)
	if len(rs) == 0 {
		return s
	}
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}
