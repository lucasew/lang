package gl

// Field/match twin of org.languagetool.rules.gl.GalicianRedundancyRule
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

// Match twin of Java example: wrong "duna de area" → fixed "duna".
func TestGalicianRedundancyRule_Rule(t *testing.T) {
	rule := NewGalicianRedundancyRule(nil)

	// incorrect sentence (Java example embedded in prose, as prior Go match twin):
	matches := rule.Match(languagetool.AnalyzePlain("A duna de area era grande."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "duna", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Pleonasmo", matches[0].GetShortMessage())
	require.Equal(t,
		"'duna de area' é un pleonasmo. É preferible dicir <suggestion>duna</suggestion>",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule2
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok, "base rule expected without sub-rule ids")
	require.Equal(t, GalicianRedundancyRuleID, base.GetID())
	require.NotNil(t, base.GetCategory())
	require.Equal(t, "REDUNDANCY", base.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, base.GetLocQualityIssueType())
	require.Equal(t, GalicianRedundancyURL, base.GetURL())

	// correct sentence: no match
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("A duna era grande."))))
}

// Field twin of GalicianRedundancyRule constructor + overrides.
func TestGalicianRedundancyRule_Metadata(t *testing.T) {
	rule := NewGalicianRedundancyRule(nil)

	require.Equal(t, GalicianRedundancyRuleID, rule.GetID())
	require.Equal(t, "GL_REDUNDANCY_REPLACE", rule.GetID())
	require.Equal(t, "1. Pleonasmos e redundancias", rule.GetDescription())
	require.Equal(t, "Pleonasmo", rule.GetShort())
	require.Equal(t,
		"'$match' é un pleonasmo. É preferible dicir $suggestions",
		rule.GetMessage())
	require.Equal(t, " ou ", rule.GetSuggestionsSeparator())
	require.Equal(t, " ou ", rule.SuggestionsSeparator)
	require.Equal(t, "gl", rule.GetLocaleCode())
	require.Equal(t, GalicianRedundancyFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/gl/redundancies.txt"}, rule.GetFileNames())
	// Java never calls useSubRuleSpecificIds()
	require.False(t, rule.SubRuleSpecificIDs)
	// Java does not override getCaseSensitivy → CaseSensitivy.CI
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	// Java: Categories.REDUNDANCY; setLocQualityIssueType(ITSIssueType.Style)
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "REDUNDANCY", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	// Java getUrl → https://gl.wikipedia.org/wiki/Pleonasmo; not setDefaultOff
	require.Equal(t, GalicianRedundancyURL, rule.GetURL())
	require.Equal(t, "https://gl.wikipedia.org/wiki/Pleonasmo", rule.GetURL())
	require.False(t, rule.IsDefaultOff())

	// example pair: duna de area → duna
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>duna de area</marker>", inc[0].GetExample())
	require.Equal(t, []string{"duna"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>duna</marker>", corr[0].GetExample())
}
