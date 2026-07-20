package tools

import (
	"bufio"
	"io"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// --- StringTools ports (org.languagetool.tools.StringTools) ---

// ApiPrintMode ports StringTools.ApiPrintMode (XML/JSON rule-match print modes).
type ApiPrintMode int

const (
	// NORMAL_API normally output the rule matches by starting and ending the XML/JSON output on every call.
	NORMAL_API ApiPrintMode = iota
	// START_API start XML/JSON output by printing the preamble and the start of the root element.
	START_API
	// END_API end XML/JSON output by closing the root element.
	END_API
	// CONTINUE_API simply continue rule match output.
	CONTINUE_API
)

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
	// Java String.split(" ") discards trailing empty strings (limit 0).
	parts := javaSplitSpace(str)
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

// javaSplitSpace mirrors Java String.split(" ") (limit 0): trailing empties discarded.
func javaSplitSpace(str string) []string {
	parts := strings.Split(str, " ")
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func UppercaseFirstChar(str string) string {
	return changeFirstCharCase(str, true)
}

// UppercaseFirstCharLang ports StringTools.uppercaseFirstChar(str, language)
// including the Dutch "ij" → "IJ" hack.
func UppercaseFirstCharLang(str, languageShortCode string) string {
	if languageShortCode == "nl" && str != "" && strings.HasPrefix(strings.ToLower(str), "ij") {
		// Java: "IJ" + str.substring(2) — first two chars are always ASCII "ij"/"IJ"/"Ij"/"iJ".
		return "IJ" + str[len("ij"):]
	}
	return changeFirstCharCase(str, true)
}

func LowercaseFirstChar(str string) string {
	return changeFirstCharCase(str, false)
}

// ConvertToTitleCaseIteratingChars ports StringTools.convertToTitleCaseIteratingChars.
// Title-cases the first letter of each space- or hyphen-separated segment; lowercases the rest.
func ConvertToTitleCaseIteratingChars(text string) string {
	if text == "" {
		return text
	}
	var b strings.Builder
	b.Grow(len(text))
	convertNext := true
	for _, ch := range text {
		// Java Character.isSpaceChar (Zs) or '-'
		if unicode.Is(unicode.Zs, ch) || ch == '-' {
			convertNext = true
			b.WriteRune(ch)
			continue
		}
		if convertNext {
			b.WriteRune(unicode.ToTitle(ch))
			convertNext = false
		} else {
			b.WriteRune(unicode.ToLower(ch))
		}
	}
	return b.String()
}

// NormalizeNFC ports StringTools.normalizeNFC.
func NormalizeNFC(str string) string {
	return norm.NFC.String(str)
}

// PreserveCase ports StringTools.preserveCase(inputString, modelString).
func PreserveCase(inputString, modelString string) string {
	if modelString == "" {
		return inputString
	}
	if IsCapitalizedWord(modelString) {
		return UppercaseFirstChar(strings.ToLower(inputString))
	}
	if IsAllUppercase(modelString) {
		return strings.ToUpper(inputString)
	}
	return inputString
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

// IsPositiveNumber ports StringTools.isPositiveNumber —
// Character.isDigit(ch) && ch != '0'.
func IsPositiveNumber(ch rune) bool {
	return unicode.IsDigit(ch) && ch != '0'
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

// wordForSpellerRE ports StringTools.WORD_FOR_SPELLER: ^[\p{L}\d\p{P}\p{Zs}]+$
var wordForSpellerRE = regexp.MustCompile(`^[\p{L}\d\p{P}\p{Zs}]+$`)

// IsEmoji ports StringTools.isEmoji — true when UTF-16 length ≠ code-point count
// (supplementary plane / surrogates) and the token is not only letters/digits/punct/spaces.
func IsEmoji(token string) bool {
	// Java: word.length() > 1 && codePointCount != length
	uLen := utf16LenTools(token)
	if uLen <= 1 {
		return false
	}
	cps := 0
	for range token {
		cps++
	}
	if cps != uLen {
		return !wordForSpellerRE.MatchString(token)
	}
	return false
}

// IsNumericSpace ports Apache Commons StringUtils.isNumericSpace (WordRepeatRule).
// True when every character is a digit or whitespace (empty → false, matching Commons).
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

// isNumericRE ports StringTools.IS_NUMERIC: ^[\d\s\.,]*\d$
var isNumericRE = regexp.MustCompile(`^[\d\s\.,]*\d$`)

// IsNumeric ports StringTools.isNumeric.
func IsNumeric(string_ string) bool {
	return isNumericRE.MatchString(string_)
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
	parts := javaSplitSpace(str)
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

// charsNotForSpelling ports StringTools.CHARS_NOT_FOR_SPELLING: [^\p{L}\d\p{P}\p{Zs}]
var charsNotForSpelling = regexp.MustCompile(`[^\p{L}\d\p{P}\p{Zs}]`)

// StringForSpeller ports StringTools.stringForSpeller — replace non-spelling symbols
// (e.g. emoji) with same-width spaces using UTF-16 length of the match.
func StringForSpeller(s string) string {
	// Java: if length > 1 && codePointCount != length (has supplementary chars)
	if utf16LenTools(s) > 1 {
		cps := 0
		for range s {
			cps++
		}
		if cps != utf16LenTools(s) {
			s = charsNotForSpelling.ReplaceAllStringFunc(s, func(found string) string {
				return strings.Repeat(" ", utf16LenTools(found))
			})
		}
	}
	return s
}

// ReadStream ports StringTools.readStream(stream, encoding):
// line-based read, each line followed by '\n' (including after the last line).
func ReadStream(r io.Reader) (string, error) {
	// Encoding is handled by the caller (bytes already decoded to a UTF-8 Reader).
	sc := bufio.NewScanner(r)
	// Allow long lines (Java uses 4k char buffer but can accumulate).
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	var sb strings.Builder
	for sc.Scan() {
		sb.WriteString(sc.Text())
		sb.WriteByte('\n')
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// Java: Pattern.compile("[\\p{IsPunctuation}']") — entire string is one punct char (or apostrophe).
var punctuationMarkRE = regexp.MustCompile(`^[\p{P}']$`)

// IsPunctuationMark ports StringTools.isPunctuationMark.
func IsPunctuationMark(input string) bool {
	return punctuationMarkRE.MatchString(input)
}

// Java: Pattern.compile("[\\p{IsPunctuation}\\p{S}']")
var punctuationOrSymbolRE = regexp.MustCompile(`^[\p{P}\p{S}']$`)

// IsPunctuationOrSymbol ports StringTools.isPunctuationOrSymbol.
func IsPunctuationOrSymbol(input string) bool {
	return punctuationOrSymbolRE.MatchString(input)
}

// Java: Pattern.compile("[^\\p{L}]+") — entire string is one or more non-letters.
var notWordStringRE = regexp.MustCompile(`^[^\p{L}]+$`)

// IsNotWordString ports StringTools.isNotWordString.
func IsNotWordString(input string) bool {
	return notWordStringRE.MatchString(input)
}

// IsAllUppercaseList ports StringTools.isAllUppercase(List<String>).
// True when every element is all-uppercase and not every element is non-letter/punct-only.
func IsAllUppercaseList(strList []string) bool {
	isInputAllUppercase := true
	isAllNotLetters := true
	for _, s := range strList {
		isInputAllUppercase = isInputAllUppercase && IsAllUppercase(s)
		isAllNotLetters = isAllNotLetters && (IsNotWordString(s) || IsPunctuationMark(s))
	}
	return isInputAllUppercase && !isAllNotLetters
}

// trimSpecialRE ports StringTools.PATTERN: (?U)[^\p{Space}\p{Alnum}\p{Punct}]
var trimSpecialRE = regexp.MustCompile(`[^\p{Z}\p{N}\p{L}\p{P}]`)

// TrimSpecialCharacters ports StringTools.trimSpecialCharacters —
// delete characters that are not space/alnum/punct (e.g. soft hyphens).
func TrimSpecialCharacters(s string) string {
	return trimSpecialRE.ReplaceAllString(s, "")
}

// NormalizeNFKC ports StringTools.normalizeNFKC.
func NormalizeNFKC(str string) string {
	return norm.NFKC.String(str)
}

// PreserveCaseWordByWord ports StringTools.preserveCaseWordByWord.
func PreserveCaseWordByWord(inputString, modelString string) string {
	// Java split(" ", -1) keeps trailing empties.
	inputWords := strings.Split(inputString, " ")
	modelWords := strings.Split(modelString, " ")
	if len(inputWords) != len(modelWords) {
		return PreserveCase(inputString, modelString)
	}
	var result strings.Builder
	for i := range inputWords {
		if i > 0 {
			result.WriteByte(' ')
		}
		result.WriteString(PreserveCase(inputWords[i], modelWords[i]))
	}
	return result.String()
}

// IsParagraphEndSentence ports StringTools.isParagraphEnd(sentence, singleLineBreaksMarksPara).
func IsParagraphEndSentence(sentence string, singleLineBreaksMarksPara bool) bool {
	if singleLineBreaksMarksPara {
		return strings.HasSuffix(sentence, "\n") || strings.HasSuffix(sentence, "\n\r")
	}
	return strings.HasSuffix(sentence, "\n\n") ||
		strings.HasSuffix(sentence, "\n\r\n\r") ||
		strings.HasSuffix(sentence, "\r\n\r\n")
}

// GetDifference ports StringTools.getDifference — single-diff split into
// [commonStart, diff1, diff2, commonEnd] using Java char (UTF-16) indices.
func GetDifference(s1, s2 string) []string {
	if s1 == s2 {
		return []string{s1, "", "", ""}
	}
	// Operate on UTF-16 code units to match Java charAt/length/substring.
	u1 := utf16Units(s1)
	u2 := utf16Units(s2)
	l1, l2 := len(u1), len(u2)
	fromStart := 0
	for fromStart < l1 && fromStart < l2 && u1[fromStart] == u2[fromStart] {
		fromStart++
	}
	fromEnd := 0
	for fromEnd < l1 && fromEnd < l2 && u1[l1-1-fromEnd] == u2[l2-1-fromEnd] {
		fromEnd++
	}
	for fromStart > l1-fromEnd {
		fromEnd--
	}
	for fromStart > l2-fromEnd {
		fromEnd--
	}
	return []string{
		utf16ToString(u1[:fromStart]),
		utf16ToString(u1[fromStart : l1-fromEnd]),
		utf16ToString(u2[fromStart : l2-fromEnd]),
		utf16ToString(u1[l1-fromEnd : l1]),
	}
}

func utf16Units(s string) []uint16 {
	// Encode to UTF-16 as Java String does.
	var out []uint16
	for _, r := range s {
		if r <= 0xFFFF {
			out = append(out, uint16(r))
		} else {
			r -= 0x10000
			out = append(out, uint16(0xD800+(r>>10)), uint16(0xDC00+(r&0x3FF)))
		}
	}
	return out
}

func utf16ToString(u []uint16) string {
	if len(u) == 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < len(u); {
		c := u[i]
		if c >= 0xD800 && c <= 0xDBFF && i+1 < len(u) {
			low := u[i+1]
			if low >= 0xDC00 && low <= 0xDFFF {
				r := rune(c-0xD800)<<10 + rune(low-0xDC00) + 0x10000
				b.WriteRune(r)
				i += 2
				continue
			}
		}
		b.WriteRune(rune(c))
		i++
	}
	return b.String()
}

// MakeWrong ports StringTools.makeWrong — invent a wrong form for speller probes.
// Note: InterrogativeVerbFilter has a *different* private makeWrong; do not merge them.
func MakeWrong(s string) string {
	repls := []struct{ old, new string }{
		{"a", "ä"}, {"e", "ë"}, {"i", "ï"}, {"o", "ö"}, {"u", "ù"},
		{"á", "ä"}, {"é", "ë"}, {"í", "ï"}, {"ó", "ö"}, {"ú", "ù"},
		{"à", "ä"}, {"è", "ë"}, {"ì", "i"}, {"ò", "ö"},
		{"ï", "ì"}, {"ü", "ù"},
	}
	for _, r := range repls {
		if strings.Contains(s, r.old) {
			return strings.ReplaceAll(s, r.old, r.new)
		}
	}
	return s + "-"
}

// NumberOf ports StringTools.numberOf — Java: s.length() - s.replace(t, "").length()
// (UTF-16 length delta; for single-char t this equals occurrence count).
func NumberOf(s, t string) int {
	return utf16LenTools(s) - utf16LenTools(strings.ReplaceAll(s, t, ""))
}

// SplitCamelCase ports StringTools.splitCamelCase.
func SplitCamelCase(input string) []string {
	if IsAllUppercase(input) {
		return []string{input}
	}
	var word, result strings.Builder
	previousIsUppercase := false
	for _, r := range input {
		if unicode.IsUpper(r) {
			if !previousIsUppercase {
				result.WriteString(word.String())
				result.WriteByte(' ')
				word.Reset()
			}
			previousIsUppercase = true
		} else {
			previousIsUppercase = false
		}
		word.WriteRune(r)
	}
	result.WriteString(word.String())
	trimmed := strings.TrimSpace(result.String())
	if trimmed == "" {
		return []string{""}
	}
	return strings.Split(trimmed, " ")
}

// SplitDigitsAtEnd ports StringTools.splitDigitsAtEnd.
// Java uses charAt + Character.isDigit from the end; digits and LT tails are BMP,
// so iterating runes with unicode.IsDigit matches Character.isDigit for those inputs.
func SplitDigitsAtEnd(input string) []string {
	runes := []rune(input)
	lastIndex := len(runes) - 1
	for lastIndex >= 0 && unicode.IsDigit(runes[lastIndex]) {
		lastIndex--
	}
	nonDigit := string(runes[:lastIndex+1])
	digit := string(runes[lastIndex+1:])
	if nonDigit != "" && digit != "" {
		return []string{nonDigit, digit}
	}
	return []string{input}
}

// IsAnagram ports StringTools.isAnagram (Java String.length is UTF-16 code units;
// sort is on UTF-16 char array — for BMP-only LT inputs this equals rune sort).
func IsAnagram(string1, string2 string) bool {
	if utf16LenTools(string1) != utf16LenTools(string2) {
		return false
	}
	a := []rune(string1)
	b := []rune(string2)
	if len(a) != len(b) {
		return false
	}
	sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
	sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// trimLeadingTrailingRE ports StringTools.TRIM_PATTERN: ^[\s\u00A0]+|[\s\u00A0]+$
var trimLeadingTrailingRE = regexp.MustCompile(`^[\s\x{00A0}]+|[\s\x{00A0}]+$`)

// TrimLeadingAndTrailingSpaces ports StringTools.trimLeadingAndTrailingSpaces.
func TrimLeadingAndTrailingSpaces(s string) string {
	return trimLeadingTrailingRE.ReplaceAllString(s, "")
}

// EscapeForXmlAttribute ports StringTools.escapeForXmlAttribute (Guava xmlAttributeEscaper).
// Escapes < > & " ' and control chars used by Guava's escaper for attributes.
func EscapeForXmlAttribute(s string) string {
	var sb strings.Builder
	for _, c := range s {
		switch c {
		case '&':
			sb.WriteString("&amp;")
		case '<':
			sb.WriteString("&lt;")
		case '>':
			sb.WriteString("&gt;")
		case '"':
			sb.WriteString("&quot;")
		case '\'':
			sb.WriteString("&apos;")
		default:
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

// EscapeForXmlContent ports StringTools.escapeForXmlContent (Guava xmlContentEscaper).
func EscapeForXmlContent(s string) string {
	var sb strings.Builder
	for _, c := range s {
		switch c {
		case '&':
			sb.WriteString("&amp;")
		case '<':
			sb.WriteString("&lt;")
		case '>':
			sb.WriteString("&gt;")
		default:
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

// StreamToString ports StringTools.streamToString (charset already applied by Reader).
func StreamToString(r io.Reader) (string, error) {
	return ReaderToString(r)
}
