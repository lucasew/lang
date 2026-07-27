package gl

// Field/match twin of org.languagetool.rules.gl.GalicianBarbarismsRule
// (no dedicated Java unit test; example pair + metadata from the rule class).
//
// Analysis: AbstractSimpleReplaceRule2 only inspects token surfaces, whitespaceBefore,
// and positions — not POS tags or lemmas. AnalyzePlain is sufficient (same surface path).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match twin of Java example: wrong "curriculum vitae" → fixed "currículo".
// barbarisms.txt: curriculum vitae=currículo
func TestGalicianBarbarismsRule_Rule(t *testing.T) {
	rule := NewGalicianBarbarismsRule(nil)

	// incorrect sentence (Java example embedded in prose, as prior Go match twin):
	matches := rule.Match(languagetool.AnalyzePlain("O curriculum vitae está completo."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "currículo", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Xenismo", matches[0].GetShortMessage())
	require.Equal(t,
		"'curriculum vitae' é un xenismo. É preferíbel dicir <suggestion>currículo</suggestion>",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule2
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok, "base rule expected without sub-rule ids")
	require.Equal(t, GalicianBarbarismsRuleID, base.GetID())
	require.NotNil(t, base.GetCategory())
	require.Equal(t, "STYLE", base.GetCategory().GetID().String())
	require.Equal(t, rules.ITSLocaleViolation, base.GetLocQualityIssueType())
	require.Equal(t, GalicianBarbarismsURL, base.GetURL())

	// correct sentence: no match
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("O currículo está completo."))))
}

// Field twin of GalicianBarbarismsRule constructor + overrides.
func TestGalicianBarbarismsRule_Metadata(t *testing.T) {
	rule := NewGalicianBarbarismsRule(nil)

	require.Equal(t, GalicianBarbarismsRuleID, rule.GetID())
	require.Equal(t, "GL_BARBARISM_REPLACE", rule.GetID())
	require.Equal(t, "Palabras de orixe estranxeira evitábeis", rule.GetDescription())
	require.Equal(t, "Xenismo", rule.GetShort())
	require.Equal(t,
		"'$match' é un xenismo. É preferíbel dicir $suggestions",
		rule.GetMessage())
	require.Equal(t, " ou ", rule.GetSuggestionsSeparator())
	require.Equal(t, " ou ", rule.SuggestionsSeparator)
	require.Equal(t, "gl", rule.GetLocaleCode())
	require.Equal(t, GalicianBarbarismsFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/gl/barbarisms.txt"}, rule.GetFileNames())
	// Java never calls useSubRuleSpecificIds()
	require.False(t, rule.SubRuleSpecificIDs)
	// Java does not override getCaseSensitivy → CaseSensitivy.CI
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	// Java: Categories.STYLE; setLocQualityIssueType(ITSIssueType.LocaleViolation)
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSLocaleViolation, rule.GetLocQualityIssueType())
	// Java getUrl → https://gl.wikipedia.org/wiki/Xenismo; not setDefaultOff
	require.Equal(t, GalicianBarbarismsURL, rule.GetURL())
	require.Equal(t, "https://gl.wikipedia.org/wiki/Xenismo", rule.GetURL())
	require.False(t, rule.IsDefaultOff())

	// example pair: curriculum vitae → currículo
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>curriculum vitae</marker>", inc[0].GetExample())
	require.Equal(t, []string{"currículo"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>currículo</marker>", corr[0].GetExample())
}
