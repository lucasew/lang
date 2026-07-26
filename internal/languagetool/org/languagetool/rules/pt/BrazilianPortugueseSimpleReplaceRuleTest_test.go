package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/BrazilianPortugueseSimpleReplaceRuleTest.java
//
// Java setUp uses JLanguageTool(Portuguese.getInstance()) +
// BrazilianPortugueseReplaceRule(messages, "/pt/pt-BR/replace.txt", language).
// AnalyzePlain is equivalent for surface matching on the replace list.
// Named-entity sentences (António / Jerónimo) are keys in official replace.txt;
// Java gets 0 matches because full analysis tags them NP*. We inject NP tags
// (no surface-name invent) so isTokenException matches Java outcomes.

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Twin of BrazilianPortugueseSimpleReplaceRuleTest.testRule
func TestBrazilianPortugueseSimpleReplaceRule_Rule(t *testing.T) {
	// Java: new BrazilianPortugueseReplaceRule(TestTools.getMessages("pt"),
	//   "/pt/pt-BR/replace.txt", lt.getLanguage())
	rule := NewBrazilianPortugueseReplaceRule(nil, BrazilianPortugueseReplaceFilePath)

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Fui de ônibus até o açougue italiano."))))

	// incorrect sentences:
	checkBrazilianPortugueseSimpleReplace(t, rule, "Vou de autocarro.", "ônibus")
	checkBrazilianPortugueseSimpleReplace(t, rule, "O lançamento de dardo é um desporto.", "esporte")
	checkBrazilianPortugueseSimpleReplace(t, rule, "Está no meu ADN!", "DNA")

	// dialect-dependent but marked as exceptions (named entities) via NP tags
	// Java: assertNoMatches("José António Miranda Coutinho")
	// Java: assertNoMatches("Jerónimo Soares")
	require.Equal(t, 0, len(rule.Match(withBRReplaceNP("José António Miranda Coutinho", "António"))))
	require.Equal(t, 0, len(rule.Match(withBRReplaceNP("Jerónimo Soares", "Jerónimo"))))
}

// checkBrazilianPortugueseSimpleReplace ports
// BrazilianPortugueseSimpleReplaceRuleTest.checkSimpleReplaceRule.
func checkBrazilianPortugueseSimpleReplace(t *testing.T, rule *BrazilianPortugueseReplaceRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// withBRReplaceNP injects POS NP* on surface (Java full PT analysis tags named entities NP).
// No surface-name invent of António/Jerónimo as hardcoded skips.
func withBRReplaceNP(text, surface string) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzePlain(text)
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil || !strings.EqualFold(tok.GetToken(), surface) {
			continue
		}
		pos := "NP00SP0"
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, nil), "test")
	}
	return sent
}

// Field twin of BrazilianPortugueseReplaceRule constructor + overrides
// (not all in Java test class, but required for sector field-by-field parity).
func TestBrazilianPortugueseSimpleReplaceRule_Metadata(t *testing.T) {
	rule := NewBrazilianPortugueseReplaceRule(nil, BrazilianPortugueseReplaceFilePath)

	require.Equal(t, BrazilianPortugueseSimpleReplaceID, rule.GetID())
	require.Equal(t, "PT_BR_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "Palavras portuguesas facilmente confundidas com as do Brasil", rule.GetDescription())
	require.Equal(t, "Palavra do português de Portugal", rule.GetShort())
	require.Equal(t,
		"\"$match\" é uma expressão usada sobretudo em Portugal. No português brasileiro diz-se $suggestions",
		rule.GetMessage())
	require.Equal(t, " ou ", rule.GetSuggestionsSeparator())
	require.Equal(t, "pt", rule.GetLocaleCode())
	require.Equal(t, []string{BrazilianPortugueseReplaceFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/pt/pt-BR/replace.txt"}, rule.GetFileNames())
	require.True(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	// Java has no getUrl override
	require.Equal(t, "", rule.GetURL())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	// Java: setLocQualityIssueType(ITSIssueType.LocaleViolation)
	require.Equal(t, rules.ITSLocaleViolation, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>hospedeira de bordo</marker>", inc[0].GetExample())
	require.Equal(t, []string{"aeromoça"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>aeromoça</marker>", corr[0].GetExample())

	// isTokenException wired (NP* || immunized)
	require.NotNil(t, rule.IsTokenException)
}

// Fail-closed: without NP tags, names that are also PT/BR pairs still match
// (no surface-name invent path). Official replace.txt contains António=Antônio.
func TestBrazilianPortugueseSimpleReplaceRule_FailClosedWithoutNP(t *testing.T) {
	rule := NewBrazilianPortugueseReplaceRule(nil, BrazilianPortugueseReplaceFilePath)
	matches := rule.Match(languagetool.AnalyzePlain("José António Miranda Coutinho"))
	require.GreaterOrEqual(t, len(matches), 1,
		"without NP tag, António must match as a PT-BR replace pair (no invent skip)")
}
