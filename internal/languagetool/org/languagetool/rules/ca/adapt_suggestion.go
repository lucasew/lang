package ca

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Catalan adaptSuggestion regexes from org.languagetool.language.Catalan.
// Go RE2 lacks (?!); caRemoveSpaces uses a trailing group instead of (?![”]).
var (
	caContractions = regexp.MustCompile(`(?i)\b([Aa]|[DdPp]e)r? e(ls?)\b`)
	caApostrophes1 = regexp.MustCompile(`\b([LDNSTMldnstm]['’]) `)
	caApostrophes2 = regexp.MustCompile(`\b([mtlsn])['’]([^1haeiouáàèéíòóúA-ZÀÈÉÍÒÓÚ“«"])`)
	caApostrophes3 = regexp.MustCompile(`(?i)\be?([mtsldn])e? (h[aeio]|h?[aeiouàèéíòóú][a-zàèéíòóúïüç])`)
	caApostrophes4 = regexp.MustCompile(`(?i)\b(l)a ([aeoàúèéí][^ ])`)
	caApostrophes5 = regexp.MustCompile(`(?i)\b([mts]e) (['’])`)
	caApostrophes6 = regexp.MustCompile(`(?i)\bs'e(ns|ls)\b`)
	caApostrophes7 = regexp.MustCompile(`(?i)\b(de|a)l (h?[aeoàúèéí][^ ])`)
	caApostrophes8 = regexp.MustCompile(`\b([MTLSN])['’]([^1haeiouáàèéíòóúA-ZÀÈÉÍÒÓÚ“«"])`)
	caApostrophes9 = regexp.MustCompile(`\b([Dd])['’]([^1haeiouáàèéíòóúA-ZÀÈÉÍÒÓÚ“«"])`)
	// Java: \b(a|de|pe) (ls?)(?!['’])\b — avoid contracting before apostrophe.
	caRemoveSpaces = regexp.MustCompile(`(?i)\b(a|de|pe) (ls?)([^\p{L}'’]|$)`)
)

// AdaptSuggestion ports Catalan.adaptSuggestion(s, originalErrorStr).
func AdaptSuggestion(s, originalErrorStr string) string {
	capitalized := tools.IsCapitalizedWord(s)
	s = strings.ReplaceAll(s, "gens traça", "gens de traça")
	s = strings.ReplaceAll(s, "gens facilitat", "gens de facilitat")
	s = caContractions.ReplaceAllString(s, "$1$2")
	s = caApostrophes1.ReplaceAllString(s, "$1")
	s = caApostrophes2.ReplaceAllString(s, "e$1 $2")
	if !strings.Contains(s, "en el") && !strings.Contains(s, "-se") {
		s = caApostrophes3.ReplaceAllString(s, "$1'$2")
	}
	s = caApostrophes4.ReplaceAllString(s, "$1'$2")
	s = caApostrophes5.ReplaceAllString(s, "$1$2")
	s = caApostrophes6.ReplaceAllString(s, "se'$1")
	s = caApostrophes7.ReplaceAllString(s, "$1 l'$2")
	// T'comença → Et comença
	s = caApostrophes8.ReplaceAllStringFunc(s, func(m string) string {
		sub := caApostrophes8.FindStringSubmatch(m)
		if len(sub) < 3 {
			return m
		}
		return "E" + strings.ToLower(sub[1]) + " " + sub[2]
	})
	s = caApostrophes9.ReplaceAllString(s, "$1e $2")
	s = caRemoveSpaces.ReplaceAllString(s, "$1$2$3")
	if capitalized {
		s = tools.UppercaseFirstChar(s)
	}
	s = strings.ReplaceAll(s, " ,", ",")
	return tools.PreserveCase(s, originalErrorStr)
}
