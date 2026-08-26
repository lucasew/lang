package gl

// Field/match twin of org.languagetool.rules.gl.GalicianWikipediaRule
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

// Match twin of Java example: wrong "a efectos de" → fixed "para os efectos de".
// wikipedia.txt: a efectos de=para os efectos de
func TestGalicianWikipediaRule_Rule(t *testing.T) {
	rule := NewGalicianWikipediaRule(nil)

	// incorrect sentence (Java example embedded in prose):
	matches := rule.Match(languagetool.AnalyzePlain("Isto vale a efectos de demonstración."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "para os efectos de", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Erro gramatical ou de normativa", matches[0].GetShortMessage())
	require.Equal(t,
		"'a efectos de' é un erro. Considere utilizar <suggestion>para os efectos de</suggestion>",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule2
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok, "base rule expected without sub-rule ids")
	require.Equal(t, GalicianWikipediaCommonErrorsID, base.GetID())
	require.NotNil(t, base.GetCategory())
	require.Equal(t, "WIKIPEDIA", base.GetCategory().GetID().String())
	require.Equal(t, rules.ITSGrammar, base.GetLocQualityIssueType())
	require.Equal(t, GalicianWikipediaURL, base.GetURL())

	// correct sentence: no match
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Isto vale para os efectos de demonstración."))))
}

// Field twin of GalicianWikipediaRule constructor + overrides.
func TestGalicianWikipediaRule_Metadata(t *testing.T) {
	rule := NewGalicianWikipediaRule(nil)

	require.Equal(t, GalicianWikipediaCommonErrorsID, rule.GetID())
	require.Equal(t, "GL_WIKIPEDIA_COMMON_ERRORS", rule.GetID())
	require.Equal(t, "Erros frecuentes nos artigos da Wikipedia", rule.GetDescription())
	require.Equal(t, "Erro gramatical ou de normativa", rule.GetShort())
	require.Equal(t,
		"'$match' é un erro. Considere utilizar $suggestions",
		rule.GetMessage())
	require.Equal(t, " ou ", rule.GetSuggestionsSeparator())
	require.Equal(t, " ou ", rule.SuggestionsSeparator)
	require.Equal(t, "gl", rule.GetLocaleCode())
	require.Equal(t, GalicianWikipediaFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/gl/wikipedia.txt"}, rule.GetFileNames())
	// Java never calls useSubRuleSpecificIds()
	require.False(t, rule.SubRuleSpecificIDs)
	// Java does not override getCaseSensitivy → CaseSensitivy.CI
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	// Java: Categories.WIKIPEDIA; setLocQualityIssueType(ITSIssueType.Grammar)
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "WIKIPEDIA", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSGrammar, rule.GetLocQualityIssueType())
	// Java getUrl → percent-encoded Erros_de_ortografía_e_desviacións; not setDefaultOff
	require.Equal(t, GalicianWikipediaURL, rule.GetURL())
	require.Equal(t,
		"https://gl.wikipedia.org/wiki/Wikipedia:Erros_de_ortograf%C3%ADa_e_desviaci%C3%B3ns",
		rule.GetURL())
	require.False(t, rule.IsDefaultOff())

	// example pair: a efectos de → para os efectos de
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>a efectos de</marker>", inc[0].GetExample())
	require.Equal(t, []string{"para os efectos de"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>para os efectos de</marker>", corr[0].GetExample())
}
