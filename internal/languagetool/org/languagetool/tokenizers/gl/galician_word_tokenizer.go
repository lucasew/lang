package gl

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// GalicianWordTokenizer ports org.languagetool.tokenizers.gl.GalicianWordTokenizer.
type GalicianWordTokenizer struct{}

func NewGalicianWordTokenizer() *GalicianWordTokenizer { return &GalicianWordTokenizer{} }

// Exact Java GalicianWordTokenizer.SPLIT_CHARS literal (character-for-character).
const splitChars = "\u0020\u002d\u00A0" +
	"\u115f\u1160\u1680" +
	"\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008" +
	"\u2009\u2013\u2014\u2015\u200A\u200B\u200c\u200d\u200e" +
	"\u200f\u2028\u2029\u202a\u202b\u202c\u202d\u202e\u202f" +
	"\u205F\u2060\u2061\u2062\u2063\u206A\u206b\u206c\u206d" +
	"\u206E\u206F\u3000\u3164\ufeff\uffa0\ufff9\ufffa\ufffb" +
	"\u002A\u002B×∗·÷:=≠≂≃≄≅≆≇≈≉≤≥≪≫∧∨∩∪∈∉∊∋∌∍" +
	",.;<>()[]{}¿¡!?:\"«»`'’‘„“”…\\/\t\r\n"

const (
	decimalCommaSubst     = '\uE001' // hide comma in decimal number temporarily
	nonBreakingSpaceSubst = '\uE002'
	nonBreakingDotSubst   = '\uE003' // hide dot in date/number temporarily
	nonBreakingColonSubst = '\uE004'
)

var (
	// Java: Pattern.CASE_INSENSITIVE|Pattern.UNICODE_CASE; \d without UNICODE_CHARACTER_CLASS = ASCII.
	decimalCommaPattern = regexp.MustCompile(`(?i)([\d]),([\d])`)
	dottedNumbers       = regexp.MustCompile(`(?i)([\d])\.([\d])`)
	colonNumbers        = regexp.MustCompile(`(?i)([\d]):([\d])`)
	// Java DATE_PATTERN — DATE_PATTERN_REPL uses only $1 $2 $3 (alts 2–3 leave empty groups; bug-for-bug).
	datePattern = regexp.MustCompile(`(?i)([\d]{2})\.([\d]{2})\.([\d]{4})|([\d]{4})\.([\d]{2})\.([\d]{2})|([\d]{4})-([\d]{2})-([\d]{2})`)
	// Java: ([\\d])\\.([aoªº][sˢ]?)
	dottedOrdinals = regexp.MustCompile(`(?i)([\d])\.([aoªº][sˢ]?)`)
	// Java DECIMAL_SPACE_PATTERN body without lookaround:
	// \d{1,3}( [\d]{3})+
	decimalSpaceBody = regexp.MustCompile(`\d{1,3}(?: \d{3})+`)
)

// Tokenize ports GalicianWordTokenizer.tokenize.
func (w *GalicianWordTokenizer) Tokenize(text string) []string {
	if strings.Contains(text, ",") {
		text = decimalCommaPattern.ReplaceAllString(text, "$1"+string(decimalCommaSubst)+"$2")
	}

	// if period is not the last character in the sentence
	// Java: indexOf('.') / length() — for ASCII '.' byte index + len is equivalent.
	dotIndex := strings.IndexByte(text, '.')
	dotInsideSentence := dotIndex >= 0 && dotIndex < len(text)-1
	if dotInsideSentence {
		// Java DATE_PATTERN_REPL = "$1"+DOT+"$2"+DOT+"$3" (only first-alt groups).
		text = datePattern.ReplaceAllString(text, "$1"+string(nonBreakingDotSubst)+"$2"+string(nonBreakingDotSubst)+"$3")
		text = dottedNumbers.ReplaceAllString(text, "$1"+string(nonBreakingDotSubst)+"$2")
		text = dottedOrdinals.ReplaceAllString(text, "$1"+string(nonBreakingDotSubst)+"$2")
	}

	// 2 000 000 — Java DECIMAL_SPACE_PATTERN + appendReplacement loop
	text = protectGLSpacedNumbers(text)

	// 12:25
	if strings.Contains(text, ":") {
		text = colonNumbers.ReplaceAllString(text, "$1"+string(nonBreakingColonSubst)+"$2")
	}

	raw := splitKeepDelims(text, splitChars)
	tokenList := make([]string, 0, len(raw))
	for _, token := range raw {
		token = strings.ReplaceAll(token, string(decimalCommaSubst), ",")
		token = strings.ReplaceAll(token, string(nonBreakingColonSubst), ":")
		token = strings.ReplaceAll(token, string(nonBreakingSpaceSubst), " ")
		// outside of if as we also replace back sentence-ending abbreviations
		token = strings.ReplaceAll(token, string(nonBreakingDotSubst), ".")
		tokenList = append(tokenList, token)
	}
	return tokenizers.JoinEMailsAndUrls(tokenList)
}

// protectGLSpacedNumbers ports Java DECIMAL_SPACE_PATTERN + Matcher appendReplacement loop.
// Java: (?<=^|[\s(])\d{1,3}( [\d]{3})+(?=[\s(]|$)
// RE2 has no lookaround; boundaries and quantifier backtracking checked manually.
// Java \s without UNICODE_CHARACTER_CLASS = ASCII [ \t\n\x0B\f\r].
func protectGLSpacedNumbers(text string) string {
	if !decimalSpaceBody.MatchString(text) {
		return text
	}
	var b strings.Builder
	i := 0
	for i < len(text) {
		loc := decimalSpaceBody.FindStringIndex(text[i:])
		if loc == nil {
			b.WriteString(text[i:])
			break
		}
		start := i + loc[0]
		fullEnd := i + loc[1]
		b.WriteString(text[i:start])

		if !glSpacedLeftOK(text, start) {
			// Cannot match at start; progress one byte so we do not re-find the same span.
			b.WriteByte(text[start])
			i = start + 1
			continue
		}

		// Java greedy ( [\d]{3})+ with backtracking: try longest end first.
		matched := false
		for _, end := range glSpacedCandidateEnds(text[start:fullEnd]) {
			absEnd := start + end
			if glSpacedRightOK(text, absEnd) {
				splitNumber := text[start:absEnd]
				// Java: replace(' ', SUBST) then replace('\u00A0', SUBST) on group(0)
				splitNumber = strings.ReplaceAll(splitNumber, " ", string(nonBreakingSpaceSubst))
				splitNumber = strings.ReplaceAll(splitNumber, "\u00A0", string(nonBreakingSpaceSubst))
				b.WriteString(splitNumber)
				i = absEnd
				matched = true
				break
			}
		}
		if !matched {
			b.WriteByte(text[start])
			i = start + 1
		}
	}
	return b.String()
}

func glSpacedLeftOK(text string, start int) bool {
	if start == 0 {
		return true
	}
	prev, _ := utf8.DecodeLastRuneInString(text[:start])
	return prev == '(' || prev == ' ' || prev == '\t' || prev == '\n' ||
		prev == '\v' || prev == '\f' || prev == '\r'
}

func glSpacedRightOK(text string, end int) bool {
	if end == len(text) {
		return true
	}
	next, _ := utf8.DecodeRuneInString(text[end:])
	return next == '(' || next == ' ' || next == '\t' || next == '\n' ||
		next == '\v' || next == '\f' || next == '\r'
}

// glSpacedCandidateEnds returns end offsets (relative to s) for s matching
// \d{1,3}(?: \d{3})+, longest first — Java greedy + backtracking order.
func glSpacedCandidateEnds(s string) []int {
	// leading 1–3 digits
	i := 0
	for i < len(s) && i < 3 && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 {
		return nil
	}
	var groupEnds []int
	for i+4 <= len(s) && s[i] == ' ' &&
		s[i+1] >= '0' && s[i+1] <= '9' &&
		s[i+2] >= '0' && s[i+2] <= '9' &&
		s[i+3] >= '0' && s[i+3] <= '9' {
		i += 4
		groupEnds = append(groupEnds, i)
	}
	// reverse for longest-first
	for l, r := 0, len(groupEnds)-1; l < r; l, r = l+1, r-1 {
		groupEnds[l], groupEnds[r] = groupEnds[r], groupEnds[l]
	}
	return groupEnds
}

// splitKeepDelims ports Java StringTokenizer(text, delims, true) for BMP delims.
func splitKeepDelims(text, delims string) []string {
	if text == "" {
		return nil
	}
	var out []string
	var cur strings.Builder
	for _, r := range text {
		if strings.ContainsRune(delims, r) {
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
			out = append(out, string(r))
		} else {
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}
