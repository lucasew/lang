package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PortugueseBarbarismRuleTest.java
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(BrazilianPortuguese).getAnalyzedSentence for this rule’s surface matching
// on the three exception sentences. None of those surfaces appear as keys in
// official /pt/pt-BR/barbarisms.txt (verified), so NP / _english_ignore_ tags from
// full PT analysis are not required for the Java-visible zero-match outcomes.
// isTokenException (immunized / NP* / _english_ignore_*) is still wired on the rule
// as in Java PortugueseBarbarismsRule.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Twin of PortugueseBarbarismRuleTest.testReplaceBarbarisms
func TestPortugueseBarbarismRule_ReplaceBarbarisms(t *testing.T) {
	// Java: new PortugueseBarbarismsRule(TestTools.getMessages("pt"), "/pt/pt-BR/barbarisms.txt",
	//       lt.getLanguage()) with lt = new JLanguageTool(BrazilianPortuguese.getInstance())
	rule := NewPortugueseBarbarismsRule(nil, PortugueseBarbarismsBRFilePath)

	// These should not have matches, since they are *exceptions* (named entities, known multi-token matches, etc.)
	// Java: assertNoMatches(...)
	assertPortugueseBarbarismNoMatches(t, rule, "New York Stock Exchange")
	assertPortugueseBarbarismNoMatches(t, rule, "Yankee Doodle, faça o morra")
	assertPortugueseBarbarismNoMatches(t, rule, "mas inferior ao Opera Browser.")
}

// assertPortugueseBarbarismNoMatches ports PortugueseBarbarismRuleTest.assertNoMatches.
func assertPortugueseBarbarismNoMatches(t *testing.T, rule *PortugueseBarbarismsRule, sentence string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 0, len(matches), "expected 0 matches for sentence: %s", sentence)
}

// Field twin of PortugueseBarbarismsRule constructor + overrides (not all in Java test class,
// but required for sector field-by-field parity — same bar as PortugueseClicheRule_Metadata).
func TestPortugueseBarbarismsRule_Metadata(t *testing.T) {
	rule := NewPortugueseBarbarismsRule(nil, PortugueseBarbarismsBRFilePath)

	require.Equal(t, PortugueseBarbarismsReplaceID, rule.GetID())
	require.Equal(t, "PT_BARBARISMS_REPLACE", rule.GetID())
	require.Equal(t, "Palavras de origem estrangeira evitáveis: $match", rule.GetDescription())
	require.Equal(t, "Estrangeirismo", rule.GetShort())
	require.Equal(t,
		"\"$match\" é um estrangeirismo. É preferível dizer $suggestions.",
		rule.GetMessage())
	require.Equal(t, " ou ", rule.GetSuggestionsSeparator())
	require.Equal(t, "pt", rule.GetLocaleCode())
	require.Equal(t, []string{PortugueseBarbarismsBRFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/pt/pt-BR/barbarisms.txt"}, rule.GetFileNames())
	require.True(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	require.Equal(t, PortugueseBarbarismsURL, rule.GetURL())
	require.Equal(t,
		"https://pt.wikipedia.org/wiki/Estrangeirismo",
		rule.GetURL())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	// Java: setLocQualityIssueType(ITSIssueType.LocaleViolation) — not Style
	require.Equal(t, rules.ITSLocaleViolation, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>curriculum vitae</marker>", inc[0].GetExample())
	require.Equal(t, []string{"currículo"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>currículo</marker>", corr[0].GetExample())

	// isTokenException wired (Java override present even when surface list alone yields 0 matches)
	require.NotNil(t, rule.IsTokenException)
}
