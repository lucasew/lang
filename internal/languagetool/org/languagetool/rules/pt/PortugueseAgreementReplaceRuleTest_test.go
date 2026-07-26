package pt

// Twin of org.languagetool.rules.pt.PortugueseAgreementReplaceRule (no dedicated
// Java *Test class; coverage bar matches PortugueseOrthographyReplaceRule ACCEPT:
// class fields + official AOreplace.txt match outcomes).
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent for this
// rule’s surface/dictionary path. AbstractSimpleReplaceRule matches token surfaces
// against wrongWords; POS/lemmas only matter when checkLemmas finds lemma keys —
// official AOreplace.txt keys are surface forms (e.g. abstracto).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official AOreplace.txt rows.
func TestPortugueseAgreementReplaceRule_Rule(t *testing.T) {
	// Java: new PortugueseAgreementReplaceRule(messages, language)
	rule := NewPortugueseAgreementReplaceRule(nil)

	// correct (post-AO form / unrelated):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("um conceito abstrato"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nada a substituir aqui."))))

	// Java example: abstracto → abstrato
	checkPortugueseAgreementReplace(t, rule, "um conceito abstracto", "abstrato")

	// Official AOreplace.txt rows
	checkPortugueseAgreementReplace(t, rule, "uma ideia abstracta", "abstrata")
	checkPortugueseAgreementReplace(t, rule, "várias abstracções", "abstrações")
	checkPortugueseAgreementReplace(t, rule, "um acto simples", "ato")

	// Case-insensitive matching (isCaseSensitive false + Locale("pt"))
	// createRuleMatch title-cases suggestion when token is capitalized
	checkPortugueseAgreementReplace(t, rule, "um conceito Abstracto", "Abstrato")

	// Java-visible createRuleMatch extras (useSubRuleSpecificIds + message/short)
	matches := rule.Match(languagetool.AnalyzePlain("um conceito abstracto"))
	require.Len(t, matches, 1)
	require.Equal(t, "Forma do Acordo Ortográfico de 45.", matches[0].GetShortMessage())
	require.Equal(t,
		"'abstracto' é uma forma do antigo acordo ortográfico. No novo acordo ortográfico, a palavra escreve-se assim: abstrato.",
		matches[0].GetMessage())
	// StringTools.toId(getId() + "_" + originalStr, language) with language pt
	wantID := tools.ToId(PortugueseAgreementReplaceID+"_"+"abstracto", "pt")
	sr, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected (useSubRuleSpecificIds)")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "PT_AGREEMENT_REPLACE_ABSTRACTO", sr.GetID())
	// getDescription has no $match placeholder → SpecificIdRule keeps static description
	require.Equal(t, "Palavras alteradas pelo Acordo Ortográfico de 90", sr.GetDescription())
	// positions: "um conceito abstracto" → abstracto at UTF-16 offset 12..21
	require.Equal(t, 12, matches[0].GetFromPos())
	require.Equal(t, 21, matches[0].GetToPos())
}

// checkPortugueseAgreementReplace asserts one match with the given first replacement.
func checkPortugueseAgreementReplace(t *testing.T, rule *PortugueseAgreementReplaceRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
	assertPortugueseAgreementRuleID(t, matches[0])
}

// assertPortugueseAgreementRuleID: useSubRuleSpecificIds → SpecificIdRule whose id starts with base id.
func assertPortugueseAgreementRuleID(t *testing.T, match *rules.RuleMatch) {
	t.Helper()
	sr, ok := match.GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected (useSubRuleSpecificIds)")
	require.True(t, strings.HasPrefix(sr.GetID(), PortugueseAgreementReplaceID),
		"rule id %q should start with %s", sr.GetID(), PortugueseAgreementReplaceID)
}

// Field twin of PortugueseAgreementReplaceRule constructor + overrides.
func TestPortugueseAgreementReplaceRule_Metadata(t *testing.T) {
	rule := NewPortugueseAgreementReplaceRule(nil)

	require.Equal(t, PortugueseAgreementReplaceID, rule.GetID())
	require.Equal(t, "PT_AGREEMENT_REPLACE", rule.GetID())
	require.Equal(t, "Palavras alteradas pelo Acordo Ortográfico de 90", rule.GetDescription())
	require.Equal(t, "Forma do Acordo Ortográfico de 45.", rule.GetShort())
	// MessageFn with empty token/replacements (static surface for metadata probes)
	require.Equal(t,
		"'' é uma forma do antigo acordo ortográfico. No novo acordo ortográfico, a palavra escreve-se assim: .",
		rule.GetMessage())

	require.Equal(t, "pt", rule.GetLocaleCode())
	require.False(t, rule.CaseSensitive)
	require.True(t, rule.CheckLemmas)
	require.True(t, rule.SubRuleSpecificIDs)
	require.Equal(t, PortugueseAgreementReplaceURL, rule.GetURL())
	require.Equal(t,
		"https://pt.wikipedia.org/wiki/Lista_das_alterações_previstas_pelo_acordo_ortográfico_de_1990",
		rule.GetURL())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	// Official dictionary loaded (same keys as AOreplace.txt)
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, []string{"abstrato"}, ww["abstracto"])
	require.Equal(t, []string{"abstrata"}, ww["abstracta"])
	require.Equal(t, []string{"ato"}, ww["acto"])

	// Category name honors supplied ResourceBundle map
	ptMsgs := map[string]string{
		"category_typo": "Erros ortográficos",
	}
	rulePT := NewPortugueseAgreementReplaceRule(ptMsgs)
	require.Equal(t, "Erros ortográficos", rulePT.GetCategory().GetName())
	// Hardcoded PT description/short do not depend on messages
	require.Equal(t, "Palavras alteradas pelo Acordo Ortográfico de 90", rulePT.GetDescription())
	require.Equal(t, "Forma do Acordo Ortográfico de 45.", rulePT.GetShort())

	// Example pair (Java addExamplePair)
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>abstracto</marker>", inc[0].GetExample())
	require.Equal(t, []string{"abstrato"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>abstrato</marker>", corr[0].GetExample())
}

