package ga

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
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

// MorphWord ports Utils.morphWord (mutations + suffixes).
func MorphWord(in string) []*Retaggable {
	var out []*Retaggable
	// First, mutations
	mut := Demutate(in)
	if mut.GetAppendTag() == ":Len:MorphError" || mut.GetAppendTag() == ":Ecl:MorphError" || mut.GetAppendTag() == ":EclLen" {
		out = append(out, mut)
		out = append(out, NewRetaggable(mut.GetWord(), mut.GetRestrictToPos(), ":DefArt:MorphError"))
	} else if mut.GetAppendTag() != "" {
		out = append(out, mut)
	}
	// Second, suffixes
	sfx := FixSuffix(mut.GetWord())
	if sfx.GetAppendTag() != "" {
		// Java always calls setAppendTag(mut.getAppendTag()) — may append or no-op if empty.
		sfx.SetAppendTag(mut.GetAppendTag())
		out = append(out, sfx)
	}
	// TODO: prefix corrections
	// TODO: other alterations
	return out
}

// Demutate ports Utils.demutate (definite-s, lenition, then eclipsis±lenition).
// "" from un* helpers corresponds to Java null.
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

// UnEclipse ports Utils.unEclipse (returns "" for Java null).
func UnEclipse(in string) string {
	// Java: if (in.length() > 2)
	if tagging.UTF16Len(in) <= 2 {
		return ""
	}
	r := []rune(in)
	switch r[0] {
	case 'N', 'n':
		ch1 := r[1]
		if len(r) > 3 && r[1] == '-' {
			ch1 = r[2]
		}
		if ch1 == 'G' || ch1 == 'D' || IsUpperVowel(ch1) || ch1 == 'g' || ch1 == 'd' || IsLowerVowel(ch1) {
			return UnEclipseChar(in, 'n', unicode.ToLower(ch1))
		}
		return ""
	case 'B', 'b':
		if r[1] == 'p' || r[1] == 'P' || (len(r) > 3 && r[1] == '-') {
			return UnEclipseChar(in, 'b', 'p')
		}
		return UnEclipseF(in)
	case 'D', 'd':
		return UnEclipseChar(in, 'd', 't')
	case 'G', 'g':
		return UnEclipseChar(in, 'g', 'c')
	case 'M', 'm':
		return UnEclipseChar(in, 'm', 'b')
	}
	return ""
}

// UnLenite ports Utils.unLenite ("" for Java null).
// Deliberately does not check if first character is one that ought to be lenited.
func UnLenite(in string) string {
	if tagging.UTF16Len(in) < 2 {
		return ""
	}
	r := []rune(in)
	if r[1] == 'h' || r[1] == 'H' {
		return string(r[0]) + string(r[2:])
	}
	return ""
}

// UnLeniteDefiniteS ports Utils.unLeniteDefiniteS ("" for Java null).
func UnLeniteDefiniteS(in string) string {
	uppers := []string{"Ts", "T-s", "TS", "T-S", "t-S", "tS"}
	lowers := []string{"ts", "t-s"}
	for _, start := range uppers {
		if tagging.UTF16Len(in) < tagging.UTF16Len(start) {
			continue
		}
		if strings.HasPrefix(in, start) {
			return "S" + in[len(start):]
		}
	}
	for _, start := range lowers {
		if tagging.UTF16Len(in) < tagging.UTF16Len(start) {
			continue
		}
		if strings.HasPrefix(in, start) {
			return "s" + in[len(start):]
		}
	}
	return ""
}

// UnEclipseF ports Utils.unEclipseF ("" for Java null).
func UnEclipseF(in string) string {
	// Java list (including duplicates) for bug-for-bug prefix order.
	uppers := []string{"Bhf", "bhF", "Bf", "bhF", "bF", "Bh-f", "bh-F", "B-f", "bh-F", "b-F"}
	lowers := []string{"bhf", "bh-f", "bf", "b-f"}
	for _, start := range uppers {
		if tagging.UTF16Len(in) < tagging.UTF16Len(start) {
			continue
		}
		if strings.HasPrefix(in, start) {
			return "F" + in[len(start):]
		}
	}
	for _, start := range lowers {
		if tagging.UTF16Len(in) < tagging.UTF16Len(start) {
			continue
		}
		if strings.HasPrefix(in, start) {
			return "f" + in[len(start):]
		}
	}
	return ""
}

// UnEclipseChar ports Utils.unEclipseChar ("" for Java null).
func UnEclipseChar(in string, first, second rune) string {
	r := []rune(in)
	from := 2
	upperFirst := unicode.ToUpper(first)
	upperSecond := unicode.ToUpper(second)
	// Java: char retSecond = (in.charAt(0) == upperFirst) ? upperSecond : second;
	// (uses charAt(0) before length bail — empty panics in Java; we bail like length < 2)
	if len(r) < 2 {
		return ""
	}
	retSecond := second
	if r[0] == upperFirst {
		retSecond = upperSecond
	}
	// no match
	if r[0] != first && r[0] != upperFirst {
		return ""
	}
	// properly eclipsed
	if r[0] == first && (r[1] == second || r[1] == upperSecond) {
		return string(r[1:])
	}
	ch1 := r[1]
	if len(r) > 3 && r[1] == '-' {
		from++
		ch1 = r[2]
	}
	if ch1 == second || ch1 == upperSecond {
		return string(retSecond) + string(r[from:])
	}
	return ""
}

// IsUpperVowel ports Utils.isUpperVowel.
func IsUpperVowel(c rune) bool {
	switch c {
	case 'A', 'E', 'I', 'O', 'U', '\u00c1', '\u00c9', '\u00cd', '\u00d3', '\u00da':
		return true
	default:
		return false
	}
}

// IsLowerVowel ports Utils.isLowerVowel.
func IsLowerVowel(c rune) bool {
	switch c {
	case 'a', 'e', 'i', 'o', 'u', '\u00e1', '\u00e9', '\u00ed', '\u00f3', '\u00fa':
		return true
	default:
		return false
	}
}

// IsVowel ports Utils.isVowel.
func IsVowel(c rune) bool {
	return IsLowerVowel(c) || IsUpperVowel(c)
}

// IsUpperLenitable ports Utils.isUpperLenitable.
func IsUpperLenitable(c rune) bool {
	switch c {
	case 'B', 'C', 'D', 'F', 'G', 'M', 'P', 'S', 'T':
		return true
	default:
		return false
	}
}

// IsLowerLenitable ports Utils.isLowerLenitable.
func IsLowerLenitable(c rune) bool {
	switch c {
	case 'b', 'c', 'd', 'f', 'g', 'm', 'p', 's', 't':
		return true
	default:
		return false
	}
}

// IsSLenitable ports Utils.isSLenitable (second letter after initial 's').
func IsSLenitable(c rune) bool {
	switch c {
	case 'l', 'n', 'r', 'a', 'e', 'i', 'o', 'u',
		'\u00e1', '\u00e9', '\u00ed', '\u00f3', '\u00fa':
		return true
	default:
		return false
	}
}

// Lenite ports Utils.lenite.
// Note: Java uses Character.isUpperCase(1) (literal code point 1) for the H/h choice —
// always false — so outh is always "h". Twin that.
func Lenite(in string) string {
	if tagging.UTF16Len(in) < 2 {
		return in
	}
	r := []rune(in)
	// Java: (Character.isUpperCase(in.charAt(0)) && Character.isUpperCase(1)) ? "H" : "h"
	// Character.isUpperCase(1) is always false → always "h"
	outh := "h"
	if IsLowerLenitable(r[0]) || IsUpperLenitable(r[0]) {
		if r[0] == 'S' || r[0] == 's' {
			if IsSLenitable(unicode.ToLower(r[1])) {
				return string(r[0]) + outh + string(r[1:])
			}
			return in
		}
		return string(r[0]) + outh + string(r[1:])
	}
	return in
}

// Eclipse ports Utils.eclipse.
func Eclipse(in string) string {
	if in == "" {
		return in
	}
	r := []rune(in)
	if IsUpperVowel(r[0]) {
		return "n" + in
	}
	if IsLowerVowel(r[0]) {
		return "n-" + in
	}
	// Java: switch(in.toLowerCase().charAt(0))
	switch unicode.ToLower(r[0]) {
	case 'b':
		return "m" + in
	case 'c':
		return "g" + in
	case 'd', 'g':
		return "n" + in
	case 'f':
		return "bh" + in
	case 'p':
		return "b" + in
	case 't':
		return "d" + in
	default:
		return in
	}
}

// ToLowerCaseIrish ports Utils.toLowerCaseIrish (tAON → t-aon, nAON → n-aon).
func ToLowerCaseIrish(s string) string {
	// Java: s.length() > 1 uses UTF-16 length
	if tagging.UTF16Len(s) > 1 {
		r := []rune(s)
		if (r[0] == 'n' || r[0] == 't') && IsUpperVowel(r[1]) {
			return string(r[0]) + "-" + strings.ToLower(string(r[1:]))
		}
	}
	return strings.ToLower(s)
}

// StartsWithUppercase ports Utils.startsWithUppercase.
func StartsWithUppercase(s string) bool {
	if startsWithMutatedUppercase(s) {
		return true
	}
	// Java: Character.isUpperCase(s.charAt(0)); empty → IndexOutOfBounds
	r := []rune(s)
	if len(r) == 0 {
		return false
	}
	return unicode.IsUpper(r[0])
}

// IsAllUppercase ports Utils.isAllUppercase.
func IsAllUppercase(s string) bool {
	r := []rune(s)
	startFrom := 0
	if startsWithMutatedUppercase(s) {
		if strings.HasPrefix(s, "bhF") {
			startFrom = 2
		} else {
			startFrom = 1
		}
	}
	if startFrom > len(r) {
		return true
	}
	for _, c := range r[startFrom:] {
		if unicode.IsLetter(c) && unicode.IsLower(c) {
			return false
		}
	}
	return true
}

// startsWithMutatedUppercase ports Utils.startsWithMutatedUppercase (private).
func startsWithMutatedUppercase(s string) bool {
	// Consonant eclipsis
	if strings.HasPrefix(s, "mB") {
		return true
	}
	if strings.HasPrefix(s, "gC") {
		return true
	}
	if strings.HasPrefix(s, "nD") {
		return true
	}
	if strings.HasPrefix(s, "bhF") {
		return true
	}
	if strings.HasPrefix(s, "nG") {
		return true
	}
	if strings.HasPrefix(s, "bP") {
		return true
	}
	if strings.HasPrefix(s, "dT") {
		return true
	}
	r := []rune(s)
	// Vowel eclipsis
	if len(r) > 1 && r[0] == 'n' && IsUpperVowel(r[1]) {
		return true
	}
	// t-prothesis
	if len(r) > 1 && r[0] == 't' && IsUpperVowel(r[1]) {
		return true
	}
	// h-prothesis
	if len(r) > 1 && r[0] == 'h' && IsUpperVowel(r[1]) {
		return true
	}
	// definite s lenition
	if strings.HasPrefix(s, "tS") {
		return true
	}
	return false
}

func isUpperPonc(c rune) bool {
	switch c {
	case 'Ḃ', 'Ċ', 'Ḋ', 'Ḟ', 'Ġ', 'Ṁ', 'Ṗ', 'Ṡ', 'Ṫ':
		return true
	default:
		return false
	}
}

func isLowerPonc(c rune) bool {
	switch c {
	case 'ḃ', 'ċ', 'ḋ', 'ḟ', 'ġ', 'ṁ', 'ṗ', 'ṡ', 'ṫ':
		return true
	default:
		return false
	}
}

// IsPonc ports Utils.isPonc.
func IsPonc(c rune) bool {
	return isUpperPonc(c) || isLowerPonc(c)
}

// ContainsPonc ports Utils.containsPonc.
func ContainsPonc(s string) bool {
	for _, c := range s {
		if IsPonc(c) {
			return true
		}
	}
	return false
}

// unPoncChar ports Utils.unPonc(char) — maps dotted letter or returns c.
func unPoncChar(c rune) rune {
	switch c {
	case 'Ḃ':
		return 'B'
	case 'Ċ':
		return 'C'
	case 'Ḋ':
		return 'D'
	case 'Ḟ':
		return 'F'
	case 'Ġ':
		return 'G'
	case 'Ṁ':
		return 'M'
	case 'Ṗ':
		return 'P'
	case 'Ṡ':
		return 'S'
	case 'Ṫ':
		return 'T'
	case 'ḃ':
		return 'b'
	case 'ċ':
		return 'c'
	case 'ḋ':
		return 'd'
	case 'ḟ':
		return 'f'
	case 'ġ':
		return 'g'
	case 'ṁ':
		return 'm'
	case 'ṗ':
		return 'p'
	case 'ṡ':
		return 's'
	case 'ṫ':
		return 't'
	default:
		return c
	}
}

// UnPonc ports Utils.unPonc(String) — dotted consonants to modern h-lenition.
func UnPonc(s string) string {
	var sb strings.Builder
	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		if !IsPonc(rs[i]) {
			sb.WriteRune(unPoncChar(rs[i]))
		} else {
			if isLowerPonc(rs[i]) {
				sb.WriteRune(unPoncChar(rs[i]))
				sb.WriteByte('h')
			} else {
				if i < len(rs)-1 && unicode.IsUpper(rs[i+1]) {
					sb.WriteRune(unPoncChar(rs[i]))
					sb.WriteByte('H')
				} else if i == len(rs)-1 && i > 0 && unicode.IsUpper(rs[i-1]) {
					sb.WriteRune(unPoncChar(rs[i]))
					sb.WriteByte('H')
				} else {
					sb.WriteRune(unPoncChar(rs[i]))
					sb.WriteByte('h')
				}
			}
		}
	}
	return sb.String()
}

// Java Utils mathematical low-surrogate constants (char values in UTF-16 trail units).
// Full code points are U+1D400.. with high surrogate U+D835.
const (
	mathematicalBoldCapitalA              = 0xDC00
	mathematicalBoldCapitalZ              = 0xDC19
	mathematicalBoldSmallA                = 0xDC1A
	mathematicalBoldSmallZ                = 0xDC33
	mathematicalItalicCapitalA            = 0xDC34
	mathematicalItalicCapitalZ            = 0xDC4D
	mathematicalItalicSmallA              = 0xDC4E
	mathematicalItalicSmallZ              = 0xDC67
	mathematicalBoldItalicCapitalA        = 0xDC68
	mathematicalBoldItalicCapitalZ        = 0xDC81
	mathematicalBoldItalicSmallA          = 0xDC82
	mathematicalBoldItalicSmallZ          = 0xDC9B
	mathematicalScriptCapitalA            = 0xDC9C
	mathematicalScriptCapitalZ            = 0xDCB5
	mathematicalScriptSmallA              = 0xDCB6
	mathematicalScriptSmallZ              = 0xDCCF
	mathematicalBoldScriptCapitalA        = 0xDCD0
	mathematicalBoldScriptCapitalZ        = 0xDCE9
	mathematicalBoldScriptSmallA          = 0xDCEA
	mathematicalBoldScriptSmallZ          = 0xDD03
	mathematicalFrakturCapitalA           = 0xDD04
	mathematicalFrakturCapitalZ           = 0xDD1D
	mathematicalFrakturSmallA             = 0xDD1E
	mathematicalFrakturSmallZ             = 0xDD37
	mathematicalDoublestruckCapitalA      = 0xDD38
	mathematicalDoublestruckCapitalZ      = 0xDD51
	mathematicalDoublestruckSmallA        = 0xDD52
	mathematicalDoublestruckSmallZ        = 0xDD6B
	mathematicalBoldFrakturCapitalA       = 0xDD6C
	mathematicalBoldFrakturCapitalZ       = 0xDD85
	mathematicalBoldFrakturSmallA         = 0xDD86
	mathematicalBoldFrakturSmallZ         = 0xDD9F
	mathematicalSansserifCapitalA         = 0xDDA0
	mathematicalSansserifCapitalZ         = 0xDDB9
	mathematicalSansserifSmallA           = 0xDDBA
	mathematicalSansserifSmallZ           = 0xDDD3
	mathematicalSansserifBoldCapitalA     = 0xDDD4
	mathematicalSansserifBoldCapitalZ     = 0xDDED
	mathematicalSansserifBoldSmallA       = 0xDDEE
	mathematicalSansserifBoldSmallZ       = 0xDE07
	mathematicalSansserifItalicCapitalA   = 0xDE08
	mathematicalSansserifItalicCapitalZ   = 0xDE21
	mathematicalSansserifItalicSmallA     = 0xDE22
	mathematicalSansserifItalicSmallZ     = 0xDE3B
	mathematicalSansserifBoldItalicCapitalA = 0xDE3C
	mathematicalSansserifBoldItalicCapitalZ = 0xDE55
	mathematicalSansserifBoldItalicSmallA = 0xDE56
	mathematicalSansserifBoldItalicSmallZ = 0xDE6F
	mathematicalMonospaceCapitalA         = 0xDE70
	mathematicalMonospaceCapitalZ         = 0xDE89
	mathematicalMonospaceSmallA           = 0xDE8A
	mathematicalMonospaceSmallZ           = 0xDEA3
	mathematicalItalicSmallDotlessI       = 0xDEA4
	mathematicalItalicSmallDotlessJ       = 0xDEA5
	mathematicalBoldCapitalAlpha          = 0xDEA8
	mathematicalBoldCapitalOmega          = 0xDEC0
	mathematicalBoldSmallAlpha            = 0xDEC2
	mathematicalBoldSmallOmega            = 0xDEDA
	mathematicalItalicCapitalAlpha        = 0xDEE2
	mathematicalItalicCapitalOmega        = 0xDEFA
	mathematicalItalicSmallAlpha          = 0xDEFC
	mathematicalItalicSmallOmega          = 0xDF14
	mathematicalBoldItalicCapitalAlpha    = 0xDF1C
	mathematicalBoldItalicCapitalOmega    = 0xDF34
	mathematicalBoldItalicSmallAlpha      = 0xDF36
	mathematicalBoldItalicSmallOmega      = 0xDF4E
	mathematicalSansserifBoldCapitalAlpha = 0xDF56
	mathematicalSansserifBoldCapitalOmega = 0xDF6E
	mathematicalSansserifBoldSmallAlpha   = 0xDF70
	mathematicalSansserifBoldSmallOmega   = 0xDF88
	mathematicalSansserifBoldItalicCapitalAlpha = 0xDF90
	mathematicalSansserifBoldItalicCapitalOmega = 0xDFA8
	mathematicalSansserifBoldItalicSmallAlpha   = 0xDFAA
	mathematicalSansserifBoldItalicSmallOmega   = 0xDFC2
	mathematicalBoldDigitZero             = 0xDFCE
	mathematicalBoldDigitNine             = 0xDFD7
	mathematicalDoublestruckDigitZero     = 0xDFD8
	mathematicalDoublestruckDigitNine     = 0xDFE1
	mathematicalSansserifDigitZero        = 0xDFE2
	mathematicalSansserifDigitNine        = 0xDFEB
	mathematicalSansserifBoldDigitZero    = 0xDFEC
	mathematicalSansserifBoldDigitNine    = 0xDFF5
	mathematicalMonospaceDigitZero        = 0xDFF6
	mathematicalMonospaceDigitNine        = 0xDFFF
	capitalA                              = 'A'
	smallA                                = 'a'
	capitalAlpha                          = 'Α'
	smallAlpha                            = 'α'
	digitZero                             = '0'
)

// IsAllMathsChars ports Utils.isAllMathsChars (UTF-16 unit walk like Java String).
func IsAllMathsChars(s string) bool {
	units := utf16.Encode([]rune(s))
	if len(units)%2 != 0 {
		return false
	}
	for i := 0; i < len(units); i++ {
		if i%2 == 0 {
			if units[i] != 0xD835 {
				return false
			}
		} else {
			numValue := int(units[i])
			if numValue < mathematicalBoldCapitalA || numValue > mathematicalMonospaceDigitNine {
				return false
			}
		}
	}
	return true
}

// IsAllHalfWidthChars ports Utils.isAllHalfWidthChars (fullwidth Latin A–Z a–z only).
func IsAllHalfWidthChars(s string) bool {
	for _, c := range s {
		charValue := int(c)
		if charValue < int('Ａ') {
			return false
		} else if charValue > int('Ｚ') && charValue < int('ａ') {
			return false
		} else if charValue > int('ｚ') {
			return false
		}
	}
	return true
}

// HalfwidthLatinToLatin ports Utils.halfwidthLatinToLatin.
func HalfwidthLatinToLatin(s string) string {
	var sb strings.Builder
	const (
		halfwidthLatinCapitalA = 'Ａ'
		halfwidthLatinCapitalZ = 'Ｚ'
		halfwidthLatinSmallA   = 'ａ'
		halfwidthLatinSmallZ   = 'ｚ'
	)
	for _, c := range s {
		charValue := int(c)
		if charValue >= halfwidthLatinCapitalA && charValue <= halfwidthLatinCapitalZ {
			sb.WriteRune(rune(charValue - halfwidthLatinCapitalA + capitalA))
		} else if charValue >= halfwidthLatinSmallA && charValue <= halfwidthLatinSmallZ {
			sb.WriteRune(rune(charValue - halfwidthLatinSmallA + smallA))
		} else {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

// getMathsChar ports Utils.getMathsChar(char, boolean, boolean).
// c is the UTF-16 low surrogate (trail unit) as in Java char after U+D835.
func getMathsChar(c uint16, normaliseGreek, normaliseDigits bool) rune {
	numeric := int(c)
	if numeric < 0 {
		// Java throws RuntimeException; unreachable for uint16
		panic("Failed to read character")
	}
	if numeric < mathematicalBoldCapitalA {
		return rune(c)
	}
	if numeric <= mathematicalBoldCapitalZ {
		return rune(numeric - mathematicalBoldCapitalA + capitalA)
	} else if numeric <= mathematicalBoldSmallZ {
		return rune(numeric - mathematicalBoldSmallA + smallA)
	} else if numeric <= mathematicalItalicCapitalZ {
		return rune(numeric - mathematicalItalicCapitalA + capitalA)
	} else if numeric <= mathematicalItalicSmallZ {
		return rune(numeric - mathematicalItalicSmallA + smallA)
	} else if numeric <= mathematicalBoldItalicCapitalZ {
		return rune(numeric - mathematicalBoldItalicCapitalA + capitalA)
	} else if numeric <= mathematicalBoldItalicSmallZ {
		return rune(numeric - mathematicalBoldItalicSmallA + smallA)
	} else if numeric <= mathematicalScriptCapitalZ {
		return rune(numeric - mathematicalScriptCapitalA + capitalA)
	} else if numeric <= mathematicalScriptSmallZ {
		return rune(numeric - mathematicalScriptSmallA + smallA)
	} else if numeric <= mathematicalBoldScriptCapitalZ {
		return rune(numeric - mathematicalBoldScriptCapitalA + capitalA)
	} else if numeric <= mathematicalBoldScriptSmallZ {
		return rune(numeric - mathematicalBoldScriptSmallA + smallA)
	} else if numeric <= mathematicalFrakturCapitalZ {
		return rune(numeric - mathematicalFrakturCapitalA + capitalA)
	} else if numeric <= mathematicalFrakturSmallZ {
		return rune(numeric - mathematicalFrakturSmallA + smallA)
	} else if numeric <= mathematicalDoublestruckCapitalZ {
		return rune(numeric - mathematicalDoublestruckCapitalA + capitalA)
	} else if numeric <= mathematicalDoublestruckSmallZ {
		return rune(numeric - mathematicalDoublestruckSmallA + smallA)
	} else if numeric <= mathematicalBoldFrakturCapitalZ {
		return rune(numeric - mathematicalBoldFrakturCapitalA + capitalA)
	} else if numeric <= mathematicalBoldFrakturSmallZ {
		return rune(numeric - mathematicalBoldFrakturSmallA + smallA)
	} else if numeric <= mathematicalSansserifCapitalZ {
		return rune(numeric - mathematicalSansserifCapitalA + capitalA)
	} else if numeric <= mathematicalSansserifSmallZ {
		return rune(numeric - mathematicalSansserifSmallA + smallA)
	} else if numeric <= mathematicalSansserifBoldCapitalZ {
		return rune(numeric - mathematicalSansserifBoldCapitalA + capitalA)
	} else if numeric <= mathematicalSansserifBoldSmallZ {
		return rune(numeric - mathematicalSansserifBoldSmallA + smallA)
	} else if numeric <= mathematicalSansserifItalicCapitalZ {
		return rune(numeric - mathematicalSansserifItalicCapitalA + capitalA)
	} else if numeric <= mathematicalSansserifItalicSmallZ {
		return rune(numeric - mathematicalSansserifItalicSmallA + smallA)
	} else if numeric <= mathematicalSansserifBoldItalicCapitalZ {
		return rune(numeric - mathematicalSansserifBoldItalicCapitalA + capitalA)
	} else if numeric <= mathematicalSansserifBoldItalicSmallZ {
		return rune(numeric - mathematicalSansserifBoldItalicSmallA + smallA)
	} else if numeric <= mathematicalMonospaceCapitalZ {
		return rune(numeric - mathematicalMonospaceCapitalA + capitalA)
	} else if numeric <= mathematicalMonospaceSmallZ {
		return rune(numeric - mathematicalMonospaceSmallA + smallA)
	} else if numeric == mathematicalItalicSmallDotlessI {
		return 'i'
	} else if numeric == mathematicalItalicSmallDotlessJ {
		return 'j'
	} else if normaliseGreek {
		if numeric >= mathematicalBoldCapitalAlpha && numeric <= mathematicalBoldCapitalOmega {
			return rune(numeric - mathematicalBoldCapitalAlpha + capitalAlpha)
		} else if numeric >= mathematicalBoldSmallAlpha && numeric <= mathematicalBoldSmallOmega {
			return rune(numeric - mathematicalBoldSmallAlpha + smallAlpha)
		} else if numeric >= mathematicalItalicCapitalAlpha && numeric <= mathematicalItalicCapitalOmega {
			return rune(numeric - mathematicalItalicCapitalAlpha + capitalAlpha)
		} else if numeric >= mathematicalItalicSmallAlpha && numeric <= mathematicalItalicSmallOmega {
			return rune(numeric - mathematicalItalicSmallAlpha + smallAlpha)
		} else if numeric >= mathematicalBoldItalicCapitalAlpha && numeric <= mathematicalBoldItalicCapitalOmega {
			return rune(numeric - mathematicalBoldItalicCapitalAlpha + capitalAlpha)
		} else if numeric >= mathematicalBoldItalicSmallAlpha && numeric <= mathematicalBoldItalicSmallOmega {
			return rune(numeric - mathematicalBoldItalicSmallAlpha + smallAlpha)
		} else if numeric >= mathematicalSansserifBoldCapitalAlpha && numeric <= mathematicalSansserifBoldCapitalOmega {
			return rune(numeric - mathematicalSansserifBoldCapitalAlpha + capitalAlpha)
		} else if numeric >= mathematicalSansserifBoldSmallAlpha && numeric <= mathematicalSansserifBoldSmallOmega {
			return rune(numeric - mathematicalSansserifBoldSmallAlpha + smallAlpha)
		} else if numeric >= mathematicalSansserifBoldItalicCapitalAlpha && numeric <= mathematicalSansserifBoldItalicCapitalOmega {
			return rune(numeric - mathematicalSansserifBoldItalicCapitalAlpha + capitalAlpha)
		} else if numeric >= mathematicalSansserifBoldItalicSmallAlpha && numeric <= mathematicalSansserifBoldItalicSmallOmega {
			return rune(numeric - mathematicalSansserifBoldItalicSmallAlpha + smallAlpha)
		}
	} else if normaliseDigits {
		if numeric >= mathematicalBoldDigitZero && numeric <= mathematicalBoldDigitNine {
			return rune(numeric - mathematicalBoldDigitZero + digitZero)
		} else if numeric <= mathematicalDoublestruckDigitNine {
			return rune(numeric - mathematicalDoublestruckDigitZero + digitZero)
		} else if numeric <= mathematicalSansserifDigitNine {
			return rune(numeric - mathematicalSansserifDigitZero + digitZero)
		} else if numeric <= mathematicalSansserifBoldDigitNine {
			return rune(numeric - mathematicalSansserifBoldDigitZero + digitZero)
		} else if numeric <= mathematicalMonospaceDigitNine {
			return rune(numeric - mathematicalMonospaceDigitZero + digitZero)
		}
	}
	return rune(c)
}

// SimplifyMathematical ports Utils.simplifyMathematical(String) — flags false,false.
func SimplifyMathematical(s string) string {
	return SimplifyMathematicalFlags(s, false, false)
}

// SimplifyMathematicalFlags ports Utils.simplifyMathematical(String, boolean, boolean).
// Iterates UTF-16 units like Java String.charAt. Non-math units are dropped (Java behavior).
func SimplifyMathematicalFlags(s string, normaliseGreek, normaliseDigits bool) string {
	units := utf16.Encode([]rune(s))
	var out strings.Builder
	for i := 0; i < len(units); i++ {
		if units[i] == 0xD835 {
			j := i + 1
			if j < len(units) && int(units[j]) >= mathematicalBoldCapitalA &&
				int(units[j]) <= mathematicalMonospaceDigitNine {
				mapped := getMathsChar(units[j], normaliseGreek, normaliseDigits)
				// Java: if (mapped == s.charAt(j)) keep both UTF-16 units; else append mapped BMP char
				if mapped == rune(units[j]) {
					// re-emit original surrogate pair as a Go string
					pair := utf16.Decode(units[i : i+2])
					for _, r := range pair {
						out.WriteRune(r)
					}
				} else {
					out.WriteRune(mapped)
				}
			}
		}
	}
	return out.String()
}

// greekLookalikes ports Utils.greekLookalikes.
func greekLookalikes(c rune) rune {
	switch c {
	case 'Α':
		return 'A'
	case 'Β':
		return 'B'
	case 'Ε':
		return 'E'
	case 'Ζ':
		return 'Z'
	case 'Η':
		return 'H'
	case 'Ι':
		return 'I'
	case 'Κ':
		return 'K'
	case 'Μ':
		return 'M'
	case 'Ν':
		return 'N'
	case 'Ο':
		return 'O'
	case 'Ρ':
		return 'P'
	case 'Τ':
		return 'T'
	case 'Υ':
		return 'Y'
	case 'Χ':
		return 'X'
	case 'α':
		return 'a'
	case 'β':
		return 'B'
	case 'γ':
		return 'y'
	case 'δ':
		return 'd'
	case 'ε':
		return 'e'
	case 'η':
		return 'n'
	case 'ι':
		return 'i'
	case 'κ':
		return 'K'
	case 'ν':
		return 'v'
	case 'ο':
		return 'o'
	case 'ρ':
		return 'p'
	case 'τ':
		return 'T'
	case 'χ':
		return 'x'
	case 'ω':
		return 'w'
	default:
		return c
	}
}

// GreekToLatin ports Utils.greekToLatin.
func GreekToLatin(s string) string {
	var sb strings.Builder
	for _, c := range s {
		sb.WriteRune(greekLookalikes(c))
	}
	return sb.String()
}

// HasMixedGreekAndLatin ports Utils.hasMixedGreekAndLatin:
// s.matches(".*[A-Za-z].*") && s.matches(".*\\p{InGREEK}.*")
func HasMixedGreekAndLatin(s string) bool {
	return reHasASCIILetter.MatchString(s) && reHasGreek.MatchString(s)
}

// HasMixedGreekAndCyrillic ports Utils.hasMixedGreekAndCyrillic:
// s.matches(".*[A-Za-z].*") && s.matches(".*\\p{InCYRILLIC}.*")
// (Java uses Latin class for first arm — twin that quirk.)
func HasMixedGreekAndCyrillic(s string) bool {
	return reHasASCIILetter.MatchString(s) && reHasCyrillic.MatchString(s)
}

var (
	// Java Pattern.matches anchors full string; .*[A-Za-z].* on full match
	reHasASCIILetter = regexp.MustCompile(`(?s).*[A-Za-z].*`)
	// \p{InGREEK} ≈ Unicode Greek and Coptic block U+0370–U+03FF
	// Use block property when available via \x{0370}-\x{03FF}; Script Greek is broader
	// but matches Java InGREEK for BMP Greek used in tests/rules.
	reHasGreek = regexp.MustCompile(`(?s).*[\x{0370}-\x{03FF}].*`)
	// \p{InCYRILLIC} ≈ U+0400–U+04FF primary block
	reHasCyrillic = regexp.MustCompile(`(?s).*[\x{0400}-\x{04FF}].*`)
)
