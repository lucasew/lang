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

// --- remaining StringTools ports ---

var nonCharID = regexp.MustCompile(`[^A-Z\x{00c0}-\x{00D6}\x{00D8}-\x{00DE}]`)

// ToId ports StringTools.toId(input, language).
func ToId(input, languageCode string) string {
	// Java String.toUpperCase maps ß → SS; Go's strings.ToUpper does not.
	trimmed := strings.TrimSpace(input)
	trimmed = strings.ReplaceAll(trimmed, "ß", "SS")
	normalised := strings.ToUpper(trimmed)
	normalised = strings.ReplaceAll(normalised, " ", "_")
	normalised = strings.ReplaceAll(normalised, "'", "_Q_")
	if languageCode == "de" {
		normalised = strings.ReplaceAll(normalised, "Ä", "AE")
		normalised = strings.ReplaceAll(normalised, "Ü", "UE")
		normalised = strings.ReplaceAll(normalised, "Ö", "OE")
	}
	return nonCharID.ReplaceAllString(normalised, "_")
}

// AddSpace ports StringTools.addSpace.
func AddSpace(word, languageShortCode string) string {
	space := " "
	if len([]rune(word)) == 1 {
		c := []rune(word)[0]
		if languageShortCode == "fr" {
			if c == '.' || c == ',' {
				space = ""
			}
		} else {
			if c == '.' || c == ',' || c == ';' || c == ':' || c == '?' || c == '!' {
				space = ""
			}
		}
	}
	return space
}

// AsString ports StringTools.asString — nil-safe CharSequence to string.
func AsString(s *string) *string {
	return s
}

// AsStringFromValue returns pointer to s (convenience for non-null).
func AsStringFromValue(s string) string { return s }

var allTitlecaseExceptions = func() map[string]struct{} {
	lists := [][]string{
		{"of", "in", "on", "the", "a", "an", "and", "or"},
		{"e", "ou", "que", "de", "do", "dos", "da", "das", "o", "a", "os", "as", "no", "nos", "na", "nas", "ao", "aos", "à", "às"},
		{"et", "ou", "que", "qui", "de", "du", "des", "en", "le", "les", "la", "un", "une", "à", "au", "aux"},
		{"y", "e", "o", "u", "que", "el", "la", "los", "las", "un", "unos", "una", "unas", "del", "nel", "de", "en", "a", "al"},
		{"von", "in", "im", "an", "am", "vom", "und", "oder", "dass", "ob", "der", "die", "das", "dem", "den", "des", "ein", "eines", "einem", "einen", "einer", "eine", "kein", "keines", "keinem", "keinen", "keiner", "keine"},
		{"van", "in", "de", "het", "een", "en", "of"},
	}
	m := map[string]struct{}{}
	for _, list := range lists {
		for _, w := range list {
			m[w] = struct{}{}
		}
	}
	return m
}()

// LowercaseFirstCharIfCapitalized ports StringTools.lowercaseFirstCharIfCapitalized.
func LowercaseFirstCharIfCapitalized(str string) string {
	if !IsCapitalizedWord(str) {
		return str
	}
	return LowercaseFirstChar(str)
}

// TitlecaseGlobal ports StringTools.titlecaseGlobal.
func TitlecaseGlobal(str string) string {
	parts := strings.Split(str, " ")
	if len(parts) == 1 {
		return UppercaseFirstChar(str)
	}
	var out []string
	for i, part := range parts {
		if i == 0 {
			out = append(out, UppercaseFirstChar(part))
			continue
		}
		if _, ok := allTitlecaseExceptions[strings.ToLower(part)]; ok {
			out = append(out, LowercaseFirstCharIfCapitalized(part))
		} else {
			out = append(out, UppercaseFirstChar(part))
		}
	}
	return strings.Join(out, " ")
}

var charsNotForSpelling = regexp.MustCompile(`[^\p{L}\d\p{P}\p{Z}]`)

// StringForSpeller ports StringTools.stringForSpeller — replace non-spelling symbols
// (e.g. emoji) with same-width spaces using UTF-16 length of the match.
func StringForSpeller(s string) string {
	// Java: if length > 1 && codePointCount != length (has supplementary chars)
	if len(s) > 1 {
		cps := 0
		for range s {
			cps++
		}
		// codePointCount vs UTF-16 length
		if cps != utf16LenTools(s) {
			// replace each match with spaces of UTF-16 length of match
			s = charsNotForSpelling.ReplaceAllStringFunc(s, func(found string) string {
				n := utf16LenTools(found)
				if n >= 20 {
					return strings.Repeat(" ", n)
				}
				return strings.Repeat(" ", n)
			})
		}
	}
	return s
}

// ReadStream ports StringTools.readStream.
func ReadStream(r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

var punctuationMarkRE = regexp.MustCompile(`^[\p{P}']+$`)

// IsPunctuationMark ports StringTools.isPunctuationMark.
func IsPunctuationMark(input string) bool {
	return punctuationMarkRE.MatchString(input)
}
