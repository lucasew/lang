package language

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// FilterSwissGermanRuleMatches ports SwissGerman.filterRuleMatches:
//  1. super.filterRuleMatches (German AI_DE_GGEC merge)
//  2. drop AI_DE_GGEC orthography matches that only "fix" ss→ß
//  3. rewrite every suggestion ß → ss
//
// Surface for step 2 is LocalMatch.OriginalSurface (Java: sentence.substring(from,to);
// ToLocalMatches also fills OriginalErrorStr). Without surface, keep the match
// (fail-closed: do not invent drop).
func FilterSwissGermanRuleMatches(matches []languagetool.LocalMatch) []languagetool.LocalMatch {
	matches = FilterGermanRuleMatches(matches)
	if len(matches) == 0 {
		return nil
	}
	out := make([]languagetool.LocalMatch, 0, len(matches))
	for _, rm := range matches {
		// Java: skip AI_DE_GGEC orthography matches when surface has "ss" and a
		// suggestion equals surface with ss→ß (would only “fix” Swiss orthography).
		if swissAIMatchIsSSToSZOnly(rm) {
			continue
		}

		// ß → ss on all suggestions
		if len(rm.Suggestions) > 0 {
			sugs := make([]string, len(rm.Suggestions))
			for i, s := range rm.Suggestions {
				sugs[i] = strings.ReplaceAll(s, "ß", "ss")
			}
			rm.Suggestions = sugs
		}
		out = append(out, rm)
	}
	return out
}

// swissAIMatchIsSSToSZOnly ports the SwissGerman AI_DE_GGEC ss→ß skip condition.
// Java: matchingString = sentence.getText().substring(fromPos, toPos).
func swissAIMatchIsSSToSZOnly(rm languagetool.LocalMatch) bool {
	if _, ok := swissGermanAISkipIDs[rm.RuleID]; !ok {
		return false
	}
	surface := rm.OriginalSurface()
	if surface == "" || !strings.Contains(surface, "ss") {
		return false
	}
	ssToSZ := strings.ReplaceAll(surface, "ss", "ß")
	for _, sug := range rm.Suggestions {
		if sug == ssToSZ {
			return true
		}
	}
	return false
}

// swissGermanAISkipIDs are Java SwissGerman filter targets for the ss→ß drop.
var swissGermanAISkipIDs = map[string]struct{}{
	"AI_DE_GGEC_REPLACEMENT_ORTHOGRAPHY_SPELL": {},
	"AI_DE_GGEC_REPLACEMENT_ADJECTIVE_FORM":    {},
}

// SwissGermanAdvancedTypography ports SwissGerman quote characters (« »).
// Opening/closing single quotes stay like German (‚ ‘).
func SwissGermanAdvancedTypography(input string) string {
	cfg := languagetool.TypographyConfig{
		Enabled:            true,
		OpeningDoubleQuote: "«",
		ClosingDoubleQuote: "»",
		OpeningSingleQuote: "‚",
		ClosingSingleQuote: "‘",
	}
	out := languagetool.ToAdvancedTypography(input, cfg)
	// Swiss inherits German TYPOGRAPHY_PATTERN for abbreviations.
	out = germanTypographyPattern.ReplaceAllString(out, "$1\u00a0$2")
	out = germanTypographyPattern.ReplaceAllString(out, "$1\u00a0$2")
	return out
}
