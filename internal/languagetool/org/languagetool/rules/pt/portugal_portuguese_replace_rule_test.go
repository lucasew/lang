package pt

// Twin of Portuguese / PortugalPortugueseReplaceRule field + list outcomes
// (no dedicated Java PortugalPortugueseSimpleReplaceRuleTest; bar mirrors
// BrazilianPortugueseSimpleReplaceRuleTest structure for the PT-PT twin).
//
// Java wiring: PortugalPortuguese → new PortugalPortugueseReplaceRule(
//   messages, "/pt/pt-PT/replace.txt", this).
// AnalyzePlain is equivalent for surface matching on the replace list.
// Named-entity surfaces (Antônio / Jerônimo) are keys in official pt-PT/replace.txt;
// Java full analysis tags them NP*. We inject NP tags (no surface-name invent)
// so isTokenException matches Java outcomes.

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Outcome twins for official pt-PT/replace.txt keys (BR surface → PT suggestion).
func TestPortugalPortugueseReplaceRule_Rule(t *testing.T) {
	// Java: new PortugalPortugueseReplaceRule(messages, "/pt/pt-PT/replace.txt", language)
	rule := NewPortugalPortugueseReplaceRule(nil, PortugalPortugueseReplaceFilePath)

	// correct sentences (European Portuguese forms — not BR keys):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Fui de autocarro até o talho italiano."))))

	// incorrect sentences (official list keys):
	// aeromoça=hospedeira de bordo (Java example pair)
	checkPortugalPortugueseSimpleReplace(t, rule, "A aeromoça serviu o café.", "hospedeira de bordo")
	// telefone celular=telemóvel
	checkPortugalPortugueseSimpleReplace(t, rule, "Comprei um telefone celular.", "telemóvel")
	// açougue=talho
	checkPortugalPortugueseSimpleReplace(t, rule, "Comprei carne no açougue.", "talho")

	// dialect-dependent but marked as exceptions (named entities) via NP tags
	// Official pt-PT keys: Antônio=António, Jerônimo=Jerónimo
	require.Equal(t, 0, len(rule.Match(withPTPTReplaceNP("José Antônio Miranda Coutinho", "Antônio"))))
	require.Equal(t, 0, len(rule.Match(withPTPTReplaceNP("Jerônimo Soares", "Jerônimo"))))

	// immunized token is also an exception (isTokenException: NP || isImmunized)
	require.Equal(t, 0, len(rule.Match(withPTPTReplaceImmunized("A aeromoça serviu o café.", "aeromoça"))))
}

func checkPortugalPortugueseSimpleReplace(t *testing.T, rule *PortugalPortugueseReplaceRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// withPTPTReplaceNP injects POS NP* on surface (Java full PT analysis tags named entities NP).
// No surface-name invent of Antônio/Jerônimo as hardcoded skips.
func withPTPTReplaceNP(text, surface string) *languagetool.AnalyzedSentence {
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

// withPTPTReplaceImmunized marks the matching surface immunized (Java isImmunized()).
func withPTPTReplaceImmunized(text, surface string) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzePlain(text)
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil || !strings.EqualFold(tok.GetToken(), surface) {
			continue
		}
		tok.Immunize(0)
	}
	return sent
}

// Field twin of PortugalPortugueseReplaceRule constructor + overrides.
func TestPortugalPortugueseReplaceRule_Metadata(t *testing.T) {
	rule := NewPortugalPortugueseReplaceRule(nil, PortugalPortugueseReplaceFilePath)

	require.Equal(t, PortugalPortugueseSimpleReplaceID, rule.GetID())
	require.Equal(t, "PT_PT_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "Brasileirismo: 1. palavras confundidas com as de Portugal", rule.GetDescription())
	require.Equal(t, "Palavra de português do Brasil", rule.GetShort())
	require.Equal(t,
		"'$match' é uma expressão brasileira, em português de Portugal utiliza-se: $suggestions",
		rule.GetMessage())
	require.Equal(t, " ou ", rule.GetSuggestionsSeparator())
	require.Equal(t, "pt", rule.GetLocaleCode())
	require.Equal(t, []string{PortugalPortugueseReplaceFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/pt/pt-PT/replace.txt"}, rule.GetFileNames())
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
	require.Equal(t, "<marker>aeromoça</marker>", inc[0].GetExample())
	require.Equal(t, []string{"hospedeira de bordo"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>hospedeira de bordo</marker>", corr[0].GetExample())

	// isTokenException wired (NP* || immunized)
	require.NotNil(t, rule.IsTokenException)
	// Direct hook: immunized / NP exception without relying solely on Match
	atr := languagetool.AnalyzePlain("aeromoça").GetTokensWithoutWhitespace()[0]
	require.False(t, rule.IsTokenException(atr))
	atr.Immunize(0)
	require.True(t, rule.IsTokenException(atr))
	atr2 := languagetool.AnalyzePlain("Antônio").GetTokensWithoutWhitespace()[0]
	pos := "NP00SP0"
	atr2.AddReading(languagetool.NewAnalyzedToken(atr2.GetToken(), &pos, nil), "test")
	require.True(t, rule.IsTokenException(atr2))
}

// Fail-closed: without NP tags, names that are also BR/PT pairs still match
// (no surface-name invent path). Official pt-PT/replace.txt contains Antônio=António.
func TestPortugalPortugueseReplaceRule_FailClosedWithoutNP(t *testing.T) {
	rule := NewPortugalPortugueseReplaceRule(nil, PortugalPortugueseReplaceFilePath)
	matches := rule.Match(languagetool.AnalyzePlain("José Antônio Miranda Coutinho"))
	require.GreaterOrEqual(t, len(matches), 1,
		"without NP tag, Antônio must match as a BR-PT replace pair (no invent skip)")
}

// Empty path → panic (Java Objects.requireNonNull(path)).
func TestPortugalPortugueseReplaceRule_RequireNonNullPath(t *testing.T) {
	require.Panics(t, func() {
		NewPortugalPortugueseReplaceRule(nil, "")
	})
}
