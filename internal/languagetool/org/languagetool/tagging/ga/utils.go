package ga

import (
	"strings"
	"unicode"
)

// Utils ports org.languagetool.tagging.ga.Utils (mutation helpers + suffix fixes).

type suffixGuess struct {
	suffix, suffixReplacement, restrictToTags, appendTags string
}

var suffixGuesses = []suffixGuess{
	{"éaracht", "éireacht", ".*Noun.*", ":MorphError"},
	{"éarachta", "éireachta", ".*Noun.*", ":MorphError"},
	{"eamhail", "iúil", ".*Noun.*|.*Adj.*", ":MorphError"},
	{"eamhuil", "iúil", ".*Noun.*|.*Adj.*", ":MorphError"},
	{"eamhla", "iúla", ".*Noun.*|.*Adj.*", ":MorphError"},
	{"amhail", "úil", ".*Noun.*|.*Adj.*", ":MorphError"},
	{"amhuil", "úil", ".*Noun.*|.*Adj.*", ":MorphError"},
}

// FixSuffix ports Utils.fixSuffix.
func FixSuffix(in string) *Retaggable {
	for _, g := range suffixGuesses {
		if strings.HasSuffix(in, g.suffix) {
			base := in[:len(in)-len(g.suffix)]
			return NewRetaggable(base+g.suffixReplacement, g.restrictToTags, g.appendTags)
		}
	}
	return NewRetaggable(in, "", "")
}

// IsVowel reports Irish vowels including fada.
func IsVowel(c rune) bool {
	switch unicode.ToLower(c) {
	case 'a', 'e', 'i', 'o', 'u', 'á', 'é', 'í', 'ó', 'ú':
		return true
	}
	return false
}

// Lenite ports Utils.lenite (insert h after initial consonant when applicable).
func Lenite(in string) string {
	if in == "" {
		return in
	}
	r := []rune(in)
	if !canLenite(r[0]) {
		return in
	}
	// already lenited
	if len(r) > 1 && (r[1] == 'h' || r[1] == 'H') {
		return in
	}
	h := 'h'
	if unicode.IsUpper(r[0]) && (len(r) == 1 || unicode.IsUpper(r[0])) {
		// keep lowercase h as in standard Irish orthography after uppercase
		h = 'h'
	}
	return string(r[0]) + string(h) + string(r[1:])
}

func canLenite(c rune) bool {
	switch unicode.ToLower(c) {
	case 'b', 'c', 'd', 'f', 'g', 'm', 'p', 's', 't':
		return true
	}
	return false
}

// Eclipse ports Utils.eclipse (consonant eclipsis prefixes).
func Eclipse(in string) string {
	if in == "" {
		return in
	}
	r := []rune(in)
	first := unicode.ToLower(r[0])
	var pref string
	switch first {
	case 'b':
		pref = "m"
	case 'c':
		pref = "g"
	case 'd':
		pref = "n"
	case 'f':
		pref = "bh"
	case 'g':
		pref = "n"
	case 'p':
		pref = "b"
	case 't':
		pref = "d"
	default:
		if IsVowel(r[0]) {
			pref = "n-"
		} else {
			return in
		}
	}
	// preserve case of first letter of word after prefix somewhat simply
	return pref + in
}

// UnLenite removes h after first character when present.
func UnLenite(in string) string {
	if len([]rune(in)) < 2 {
		return ""
	}
	r := []rune(in)
	if r[1] == 'h' || r[1] == 'H' {
		return string(r[0]) + string(r[2:])
	}
	return ""
}

// Demutate ports Utils.demutate (definite-s, lenition, then eclipsis±lenition).
func Demutate(in string) *Retaggable {
	if out := UnLeniteDefiniteS(in); out != "" {
		return NewRetaggable(out, "(?:C[UMC]:)?Noun:.*:DefArt", ":MorphError")
	}
	if out := UnLenite(in); out != "" {
		return NewRetaggable(out, "", ":Len:MorphError")
	}
	if out := UnEclipse(in); out != "" {
		if out2 := UnLenite(out); out2 != "" {
			return NewRetaggable(out2, "", ":EclLen")
		}
		return NewRetaggable(out, "", ":Ecl:MorphError")
	}
	return NewRetaggable(in, "", "")
}

// MorphWord ports Utils.morphWord (mutations + suffixes).
func MorphWord(in string) []*Retaggable {
	var out []*Retaggable
	mut := Demutate(in)
	if mut.GetAppendTag() != "" {
		out = append(out, mut)
	}
	sfx := FixSuffix(mut.GetWord())
	if sfx.GetAppendTag() != "" {
		if mut.GetAppendTag() != "" {
			sfx.SetAppendTag(mut.GetAppendTag())
		}
		out = append(out, sfx)
	}
	return out
}

// ToLowerCaseIrish ports Utils.toLowerCaseIrish (tAON → t-aon, nAON → n-aon).
func ToLowerCaseIrish(s string) string {
	if len(s) > 1 {
		r := []rune(s)
		if (r[0] == 'n' || r[0] == 't') && isUpperVowel(r[1]) {
			return string(r[0]) + "-" + strings.ToLower(string(r[1:]))
		}
	}
	return strings.ToLower(s)
}

func isUpperVowel(c rune) bool {
	switch c {
	case 'A', 'E', 'I', 'O', 'U', 'Á', 'É', 'Í', 'Ó', 'Ú':
		return true
	}
	return false
}

// UnEclipseChar ports Utils.unEclipseChar.
func UnEclipseChar(in string, first, second rune) string {
	r := []rune(in)
	if len(r) < 2 {
		return ""
	}
	upperFirst := unicode.ToUpper(first)
	upperSecond := unicode.ToUpper(second)
	retSecond := second
	if r[0] == upperFirst {
		retSecond = upperSecond
	}
	if r[0] != first && r[0] != upperFirst {
		return ""
	}
	// properly eclipsed: first + second
	if r[0] == first && (r[1] == second || r[1] == upperSecond) {
		return string(r[1:])
	}
	from := 2
	ch1 := r[1]
	if len(r) > 3 && r[1] == '-' {
		from = 3
		ch1 = r[2]
	}
	if ch1 == second || ch1 == upperSecond {
		return string(retSecond) + string(r[from:])
	}
	return ""
}

// UnEclipse ports Utils.unEclipse (returns "" for Java null).
func UnEclipse(in string) string {
	if len([]rune(in)) <= 2 {
		return ""
	}
	r := []rune(in)
	switch r[0] {
	case 'N', 'n':
		ch1 := r[1]
		if len(r) > 3 && r[1] == '-' {
			ch1 = r[2]
		}
		if ch1 == 'G' || ch1 == 'D' || ch1 == 'g' || ch1 == 'd' || isUpperVowel(ch1) || IsVowel(ch1) {
			return UnEclipseChar(in, 'n', unicode.ToLower(ch1))
		}
		return ""
	case 'B', 'b':
		if r[1] == 'p' || r[1] == 'P' || (len(r) > 3 && r[1] == '-') {
			return UnEclipseChar(in, 'b', 'p')
		}
		return unEclipseF(in)
	case 'D', 'd':
		return UnEclipseChar(in, 'd', 't')
	case 'G', 'g':
		return UnEclipseChar(in, 'g', 'c')
	case 'M', 'm':
		return UnEclipseChar(in, 'm', 'b')
	}
	return ""
}

func unEclipseF(in string) string {
	uppers := []string{"Bhf", "bhF", "Bf", "bF", "Bh-f", "bh-F", "B-f", "b-F"}
	lowers := []string{"bhf", "bh-f", "bf", "b-f"}
	for _, start := range uppers {
		if strings.HasPrefix(in, start) {
			return "F" + in[len(start):]
		}
	}
	for _, start := range lowers {
		if strings.HasPrefix(in, start) {
			return "f" + in[len(start):]
		}
	}
	return ""
}

// UnLeniteDefiniteS ports Utils.unLeniteDefiniteS ("" for Java null).
func UnLeniteDefiniteS(in string) string {
	uppers := []string{"Ts", "T-s", "TS", "T-S", "t-S", "tS"}
	lowers := []string{"ts", "t-s"}
	for _, start := range uppers {
		if strings.HasPrefix(in, start) {
			return "S" + in[len(start):]
		}
	}
	for _, start := range lowers {
		if strings.HasPrefix(in, start) {
			return "s" + in[len(start):]
		}
	}
	return ""
}

// UnPonc converts dotted consonants (ḃ→bh) — ports Utils.unPonc.
func UnPonc(s string) string {
	var b strings.Builder
	rs := []rune(s)
	for i, c := range rs {
		base, ok := unPoncChar(c)
		if !ok {
			b.WriteRune(c)
			continue
		}
		b.WriteRune(base)
		// append h/H
		if unicode.IsLower(c) {
			b.WriteByte('h')
		} else {
			if i < len(rs)-1 && unicode.IsUpper(rs[i+1]) {
				b.WriteByte('H')
			} else if i == len(rs)-1 && i > 0 && unicode.IsUpper(rs[i-1]) {
				b.WriteByte('H')
			} else {
				b.WriteByte('h')
			}
		}
	}
	return b.String()
}

func unPoncChar(c rune) (rune, bool) {
	m := map[rune]rune{
		'Ḃ': 'B', 'ḃ': 'b',
		'Ċ': 'C', 'ċ': 'c',
		'Ḋ': 'D', 'ḋ': 'd',
		'Ḟ': 'F', 'ḟ': 'f',
		'Ġ': 'G', 'ġ': 'g',
		'Ṁ': 'M', 'ṁ': 'm',
		'Ṗ': 'P', 'ṗ': 'p',
		'Ṡ': 'S', 'ṡ': 's',
		'Ṫ': 'T', 'ṫ': 't',
	}
	base, ok := m[c]
	return base, ok
}

// SimplifyMathematical maps mathematical alphanumeric symbols (bold plane) to ASCII.
func SimplifyMathematical(s string) string {
	var out strings.Builder
	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		// surrogate pair for U+1D400.. as Go runes are already decoded
		c := rs[i]
		if c >= 0x1D400 && c <= 0x1D419 { // bold capital A-Z
			out.WriteRune('A' + (c - 0x1D400))
			continue
		}
		if c >= 0x1D41A && c <= 0x1D433 { // bold small a-z
			out.WriteRune('a' + (c - 0x1D41A))
			continue
		}
		out.WriteRune(c)
	}
	return out.String()
}
