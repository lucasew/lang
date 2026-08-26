package language

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Patterns from Catalan.java. Java uses Pattern.UNICODE_CHARACTER_CLASS so \b
// respects Unicode letters. Go RE2 \b is ASCII-only; we match without \b and
// enforce Java-like word edges via isCAWordChar (letter/mark/digit/_).
// Only enforce the same \b sides as each Java pattern (many have leading \b only).
var (
	caContractions = regexp.MustCompile(`(?i)([Aa]|[DdPp]e)r? e(ls?)`)
	caApostrophes1 = regexp.MustCompile(`([LDNSTMldnstm]['’]) `)
	caApostrophes2 = regexp.MustCompile(`([mtlsn])['’]([^1haeiouáàèéíòóúA-ZÀÈÉÍÒÓÚ“«"])`)
	caApostrophes3 = regexp.MustCompile(`(?i)e?([mtsldn])e? (h[aeio]|h?[aeiouàèéíòóú][a-zàèéíòóúïüç])`)
	caApostrophes4 = regexp.MustCompile(`(?i)(l)a ([aeoàúèéí][^ ])`)
	caApostrophes5 = regexp.MustCompile(`(?i)([mts]e) (['’])`)
	caApostrophes6 = regexp.MustCompile(`(?i)s'e(ns|ls)`)
	caApostrophes7 = regexp.MustCompile(`(?i)(de|a)l (h?[aeoàúèéí][^ ])`)
	caApostrophes8 = regexp.MustCompile(`([MTLSN])['’]([^1haeiouáàèéíòóúA-ZÀÈÉÍÒÓÚ“«"])`)
	caApostrophes9 = regexp.MustCompile(`([Dd])['’]([^1haeiouáàèéíòóúA-ZÀÈÉÍÒÓÚ“«"])`)
	// Java: \b(a|de|pe) (ls?)(?![''])\b — RE2 has no (?!); apostrophe checked in replace.
	caRemoveSpaces = regexp.MustCompile(`(?i)(a|de|pe) (ls?)`)
)

// isCAWordChar approximates Java UNICODE_CHARACTER_CLASS \w for Catalan \b edges.
func isCAWordChar(r rune) bool {
	if r == '_' {
		return true
	}
	if unicode.IsLetter(r) || unicode.IsNumber(r) {
		return true
	}
	return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) || unicode.Is(unicode.Me, r)
}

func caLeftWordBound(s string, i int) bool {
	if i <= 0 {
		return true
	}
	r, _ := utf8.DecodeLastRuneInString(s[:i])
	if r == utf8.RuneError {
		return true
	}
	return !isCAWordChar(r)
}

func caRightWordBound(s string, end int) bool {
	if end >= len(s) {
		return true
	}
	r, _ := utf8.DecodeRuneInString(s[end:])
	if r == utf8.RuneError {
		return true
	}
	return !isCAWordChar(r)
}

// replaceCABounds applies re on matches with left and/or right Unicode word bounds.
func replaceCABounds(s string, re *regexp.Regexp, repl string, left, right bool) string {
	locs := re.FindAllStringSubmatchIndex(s, -1)
	if len(locs) == 0 {
		return s
	}
	var b strings.Builder
	last := 0
	for _, loc := range locs {
		if len(loc) < 2 {
			continue
		}
		start, end := loc[0], loc[1]
		if left && !caLeftWordBound(s, start) {
			continue
		}
		if right && !caRightWordBound(s, end) {
			continue
		}
		b.WriteString(s[last:start])
		b.Write(re.ExpandString(nil, repl, s, loc))
		last = end
	}
	b.WriteString(s[last:])
	return b.String()
}

// CatalanAdaptSuggestion ports Catalan.adaptSuggestion.
func CatalanAdaptSuggestion(s, originalErrorStr string) string {
	capitalized := tools.IsCapitalizedWord(s)
	s = strings.ReplaceAll(s, "gens traça", "gens de traça")
	s = strings.ReplaceAll(s, "gens facilitat", "gens de facilitat")
	// Java CA_CONTRACTIONS: \b … \b
	s = replaceCABounds(s, caContractions, "$1$2", true, true)
	// Java CA_APOSTROPHES1–5,7–9: leading \b only
	s = replaceCABounds(s, caApostrophes1, "$1", true, false)
	s = replaceCABounds(s, caApostrophes2, "e$1 $2", true, false)
	if !strings.Contains(s, "en el") && !strings.Contains(s, "-se") {
		s = replaceCABounds(s, caApostrophes3, "$1'$2", true, false)
	}
	s = replaceCABounds(s, caApostrophes4, "$1'$2", true, false)
	s = replaceCABounds(s, caApostrophes5, "$1$2", true, false)
	// Java CA_APOSTROPHES6: \b … \b
	s = replaceCABounds(s, caApostrophes6, "se'$1", true, true)
	s = replaceCABounds(s, caApostrophes7, "$1 l'$2", true, false)
	// T'comença -> Et comença (Java appendReplacement with lowercased group1)
	s = replaceCAApostrophes8(s)
	s = replaceCABounds(s, caApostrophes9, "$1e $2", true, false)
	s = replaceCARemoveSpaces(s)
	if capitalized {
		s = tools.UppercaseFirstChar(s)
	}
	s = strings.ReplaceAll(s, " ,", ",")
	return tools.PreserveCase(s, originalErrorStr)
}

func replaceCAApostrophes8(s string) string {
	// Java: leading \b only; while find → "E" + group1.toLowerCase() + " " + group2
	var b strings.Builder
	last := 0
	for _, loc := range caApostrophes8.FindAllStringSubmatchIndex(s, -1) {
		if len(loc) < 6 {
			continue
		}
		start, end := loc[0], loc[1]
		if !caLeftWordBound(s, start) {
			continue
		}
		b.WriteString(s[last:start])
		g1 := s[loc[2]:loc[3]]
		g2 := s[loc[4]:loc[5]]
		b.WriteString("E")
		b.WriteString(strings.ToLower(g1))
		b.WriteByte(' ')
		b.WriteString(g2)
		last = end
	}
	b.WriteString(s[last:])
	return b.String()
}

// replaceCARemoveSpaces ports CA_REMOVE_SPACES with Java (?!['']) and \b…\b.
func replaceCARemoveSpaces(s string) string {
	var b strings.Builder
	last := 0
	for _, loc := range caRemoveSpaces.FindAllStringSubmatchIndex(s, -1) {
		if len(loc) < 6 {
			continue
		}
		start, end := loc[0], loc[1]
		if !caLeftWordBound(s, start) || !caRightWordBound(s, end) {
			continue
		}
		if end < len(s) {
			rest := s[end:]
			if strings.HasPrefix(rest, "'") || strings.HasPrefix(rest, "’") {
				continue
			}
		}
		b.WriteString(s[last:start])
		b.WriteString(s[loc[2]:loc[3]])
		b.WriteString(s[loc[4]:loc[5]])
		last = end
	}
	b.WriteString(s[last:])
	return b.String()
}
