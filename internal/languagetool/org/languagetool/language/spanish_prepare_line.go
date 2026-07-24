package language

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SpanishPrepareLineForSpeller ports Spanish.prepareLineForSpeller.
// After '#', form/tag split on tab or semicolon; keep form when tag is N*, _Latin_, or LOC_ADV.
func SpanishPrepareLineForSpeller(line string) []string {
	parts := strings.Split(line, "#")
	if len(parts) == 0 {
		return []string{line}
	}
	formTag := regexp.MustCompile(`[\t;]`).Split(parts[0], -1)
	if len(formTag) > 1 {
		// Java: formTag[i].trim()
		tag := tools.JavaStringTrim(formTag[1])
		form := tools.JavaStringTrim(formTag[0])
		if strings.HasPrefix(tag, "N") || tag == "_Latin_" || tag == "LOC_ADV" {
			return []string{form}
		}
		return []string{""}
	}
	return []string{line}
}

// esContractions ports Spanish.ES_CONTRACTIONS: \b([Aa]|[Dd]e) e(l)\b → $1$2
var esContractions = regexp.MustCompile(`\b([Aa]|[Dd]e) e(l)\b`)

// SpanishAdaptSuggestion ports Spanish.adaptSuggestion (a/de el → al/del).
func SpanishAdaptSuggestion(replacement, originalErrorStr string) string {
	_ = originalErrorStr // Java keeps originalErrorStr in signature; unused in body
	return esContractions.ReplaceAllString(replacement, "$1$2")
}

// SpanishHasMinMatchesRules ports Spanish.hasMinMatchesRules (true).
func SpanishHasMinMatchesRules() bool { return true }
