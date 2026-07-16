package tools

import (
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// --- StringTools ports (org.languagetool.tools.StringTools) ---

func IsEmptyStr(str string) bool { return str == "" }

func AssureSet(s, varName string) {
	if IsEmptyStr(strings.TrimSpace(s)) {
		panic("IllegalArgumentException: " + varName + " cannot be empty or whitespace only")
	}
}

func IsAllUppercase(str string) bool {
	for _, c := range str {
		if unicode.IsLetter(c) && unicode.IsLower(c) {
			return false
		}
	}
	return true
}

func IsNotAllLowercase(str string) bool {
	for _, c := range str {
		if unicode.IsLetter(c) && !unicode.IsLower(c) {
			return true
		}
	}
	return false
}

func IsCapitalizedWord(str string) bool {
	if IsEmptyStr(str) {
		return false
	}
	r, size := utf8.DecodeRuneInString(str)
	if !unicode.IsUpper(r) {
		return false
	}
	for _, c := range str[size:] {
		if unicode.IsLetter(c) && !unicode.IsLower(c) {
			return false
		}
	}
	return true
}

func IsMixedCase(str string) bool {
	return !IsAllUppercase(str) && !IsCapitalizedWord(str) && IsNotAllLowercase(str)
}

func StartsWithUppercase(str string) bool {
	if IsEmptyStr(str) {
		return false
	}
	r, _ := utf8.DecodeRuneInString(str)
	return unicode.IsUpper(r)
}

func StartsWithLowercase(str string) bool {
	if IsEmptyStr(str) {
		return false
	}
	r, _ := utf8.DecodeRuneInString(str)
	return unicode.IsLower(r)
}

func AllStartWithLowercase(str string) bool {
	parts := strings.Split(str, " ")
	if len(parts) < 2 {
		return StartsWithLowercase(str)
	}
	for _, p := range parts {
		if !StartsWithLowercase(p) {
			return false
		}
	}
	return true
}

func UppercaseFirstChar(str string) string {
	return changeFirstCharCase(str, true)
}

func LowercaseFirstChar(str string) string {
	return changeFirstCharCase(str, false)
}

func changeFirstCharCase(str string, toUpperCase bool) string {
	if IsEmptyStr(str) {
		return str
	}
	runes := []rune(str)
	if len(runes) == 1 {
		if toUpperCase {
			return strings.ToUpper(str) // Locale.ENGLISH for letters
		}
		return strings.ToLower(str)
	}
	pos := 0
	lenR := len(runes) - 1
	for !unicode.IsLetter(runes[pos]) && !unicode.IsDigit(runes[pos]) && lenR > pos {
		pos++
	}
	first := runes[pos]
	if toUpperCase {
		first = unicode.ToUpper(first)
	} else {
		first = unicode.ToLower(first)
	}
	return string(runes[:pos]) + string(first) + string(runes[pos+1:])
}

func EscapeXML(s string) string { return EscapeHTML(s) }

func EscapeHTML(s string) string {
	var sb strings.Builder
	for _, c := range s {
		switch c {
		case '<':
			sb.WriteString("&lt;")
		case '>':
			sb.WriteString("&gt;")
		case '&':
			sb.WriteString("&amp;")
		case '"':
			sb.WriteString("&quot;")
		default:
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

// TrimWhitespace ports StringTools.trimWhitespace.
func TrimWhitespace(s string) string {
	str := strings.TrimSpace(s)
	// Java uses charAt with <= ' ' for whitespace
	var filter strings.Builder
	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		for runes[i] <= ' ' && i < len(runes) &&
			(i+1 < len(runes) && runes[i+1] <= ' ' || i > 1 && runes[i-1] <= ' ') {
			i++
			if i >= len(runes) {
				break
			}
		}
		if i >= len(runes) {
			break
		}
		c := runes[i]
		if c != '\n' && c != '\t' && c != '\r' {
			filter.WriteRune(c)
		}
	}
	out := filter.String()
	if len([]rune(out)) == len(runes) {
		return str
	}
	return out
}

// IsWhitespace ports StringTools.isWhitespace.
func IsWhitespace(str string) bool {
	if str == "\u0002" || str == "\u0001" {
		return false
	}
	if str == "\uFEFF" {
		return true
	}
	trimStr := trimJava(str)
	if trimStr == "" {
		return true
	}
	if len([]rune(trimStr)) == 1 {
		if str == "\u200B" || str == "\u00A0" || str == "\u202F" {
			return true
		}
		r := []rune(trimStr)[0]
		return unicode.IsSpace(r)
	}
	return false
}

func trimJava(s string) string {
	start, end := 0, len(s)
	for start < end && s[start] <= ' ' {
		start++
	}
	for end > start && s[end-1] <= ' ' {
		end--
	}
	return s[start:end]
}

func IsPositiveNumber(ch rune) bool {
	return ch >= '1' && ch <= '9'
}

func FilterXML(str string) string {
	s := str
	if !strings.Contains(s, "<") {
		return s
	}
	s = regexp.MustCompile(`(?s)<!--.*?-->`).ReplaceAllString(s, " ")
	// Java: (?<!<)<[^<>]+>  — tag not preceded by another <
	var out strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); {
		if runes[i] == '<' && (i == 0 || runes[i-1] != '<') {
			// find closing >
			j := i + 1
			ok := true
			for j < len(runes) && runes[j] != '>' {
				if runes[j] == '<' {
					ok = false
					break
				}
				j++
			}
			if ok && j < len(runes) && runes[j] == '>' {
				i = j + 1
				continue // drop tag
			}
		}
		out.WriteRune(runes[i])
		i++
	}
	return out.String()
}

func ReaderToString(r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	return string(b), err
}

// IsCamelCase ports token.matches("[a-z]+[A-Z][A-Za-z]+")
func IsCamelCase(str string) bool {
	ok, _ := regexp.MatchString(`^[a-z]+[A-Z][A-Za-z]+$`, str)
	return ok
}

func IsNonBreakingWhitespace(str string) bool {
	return str == "\u00A0"
}

// IsEmoji — minimal: false for BMP tests; detect surrogate pairs / emoji ranges loosely
func IsEmoji(token string) bool {
	for _, r := range token {
		if r >= 0x1F300 && r <= 0x1FAFF {
			return true
		}
		if r >= 0x2600 && r <= 0x27BF {
			return true
		}
	}
	// also multi-codepoint emoji as UTF-16 surrogates already decoded to runes
	return false
}

func IsNumericSpace(token string) bool {
	if token == "" {
		return false
	}
	for _, r := range token {
		if !(unicode.IsDigit(r) || unicode.IsSpace(r)) {
			return false
		}
	}
	return true
}

// IsNotWordCharacter ports StringTools.isNotWordCharacter (Pattern "[^\p{L}]").matches.
func IsNotWordCharacter(input string) bool {
	runes := []rune(input)
	if len(runes) != 1 {
		// Java pattern is a single char class; multi-char never matches fully
		return false
	}
	return !unicode.IsLetter(runes[0])
}
