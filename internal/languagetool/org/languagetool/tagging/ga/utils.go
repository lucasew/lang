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

// Demutate attempts reverse lenition/eclipsis (simplified).
func Demutate(in string) *Retaggable {
	if un := UnLenite(in); un != "" {
		return NewRetaggable(un, "", ":Len:MorphError")
	}
	// simple eclipsis reverse for common prefixes
	lower := strings.ToLower(in)
	for _, p := range []struct{ pref, tag string }{
		{"bhf", ":Ecl:MorphError"},
		{"mb", ":Ecl:MorphError"},
		{"gc", ":Ecl:MorphError"},
		{"nd", ":Ecl:MorphError"},
		{"ng", ":Ecl:MorphError"},
		{"bp", ":Ecl:MorphError"},
		{"dt", ":Ecl:MorphError"},
		{"n-", ":Ecl:MorphError"},
	} {
		if strings.HasPrefix(lower, p.pref) {
			rest := in[len(p.pref):]
			if p.pref == "bhf" {
				rest = "f" + in[len(p.pref):]
				if unicode.IsUpper([]rune(in)[0]) {
					rest = "F" + in[len(p.pref):]
				}
			}
			return NewRetaggable(rest, "", p.tag)
		}
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
