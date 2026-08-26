package gl

// Field/match twin of org.languagetool.rules.gl.GalicianWordinessRule
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

// Match twin of Java example: wrong "Raramente é o caso en que acontece" → fixed "Raramente acontece".
// wordiness.txt: raramente é o caso en que acontece=raramente acontece
// ASR2 CI + sentence-start capitalization → suggestion "Raramente acontece".
func TestGalicianWordinessRule_Rule(t *testing.T) {
	rule := NewGalicianWordinessRule(nil)

	// incorrect sentence (Java example embedded in prose, as GL Redundancy match twin):
	matches := rule.Match(languagetool.AnalyzePlain("Raramente é o caso en que acontece isto."))
	require.Equal(t, 1, len(matches))
	// createMatch: sentStart + capitalized original → UppercaseFirstChar on suggestion
	require.Equal(t, "Raramente acontece", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Expresión prolixa", matches[0].GetShortMessage())
	// message template uses raw replacements from data (not case-adjusted)
	require.Equal(t,
		"'Raramente é o caso en que acontece' é unha expresión innecesariamente complexa. É preferíbel dicir <suggestion>raramente acontece</suggestion>",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule2
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok, "base rule expected without sub-rule ids")
	require.Equal(t, GalicianWordinessRuleID, base.GetID())
	require.NotNil(t, base.GetCategory())
	require.Equal(t, "REDUNDANCY", base.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, base.GetLocQualityIssueType())
	// Java getUrl is commented out
	require.Equal(t, "", base.GetURL())

	// correct sentence: no match
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Raramente acontece isto."))))
}

// Field twin of GalicianWordinessRule constructor + overrides.
func TestGalicianWordinessRule_Metadata(t *testing.T) {
	rule := NewGalicianWordinessRule(nil)

	require.Equal(t, GalicianWordinessRuleID, rule.GetID())
	require.Equal(t, "GL_WORDINESS_REPLACE", rule.GetID())
	require.Equal(t, "2. Expresións prolixas", rule.GetDescription())
	require.Equal(t, "Expresión prolixa", rule.GetShort())
	require.Equal(t,
		"'$match' é unha expresión innecesariamente complexa. É preferíbel dicir $suggestions",
		rule.GetMessage())
	require.Equal(t, " ou ", rule.GetSuggestionsSeparator())
	require.Equal(t, " ou ", rule.SuggestionsSeparator)
	require.Equal(t, "gl", rule.GetLocaleCode())
	require.Equal(t, GalicianWordinessFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/gl/wordiness.txt"}, rule.GetFileNames())
	// Java never calls useSubRuleSpecificIds()
	require.False(t, rule.SubRuleSpecificIDs)
	// Java does not override getCaseSensitivy → CaseSensitivy.CI
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	// Java: Categories.REDUNDANCY; setLocQualityIssueType(ITSIssueType.Style)
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "REDUNDANCY", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	// Java getUrl is COMMENTED OUT — no URL; not setDefaultOff
	require.Equal(t, "", rule.GetURL())
	require.False(t, rule.IsDefaultOff())

	// example pair: Raramente é o caso en que acontece → Raramente acontece
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>Raramente é o caso en que acontece</marker> isto.", inc[0].GetExample())
	require.Equal(t, []string{"Raramente acontece"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>Raramente acontece</marker> isto.", corr[0].GetExample())
}
