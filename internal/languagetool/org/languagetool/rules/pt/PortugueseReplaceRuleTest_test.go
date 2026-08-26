package pt

// Twin of org.languagetool.rules.pt.PortugueseReplaceRule (no dedicated
// Java *Test class; coverage bar matches PortugueseAgreementReplaceRule ACCEPT:
// class fields + official /pt/replace.txt load/match outcomes).
//
// Official replace.txt is comments-only → empty wrongWords map → no matches
// on arbitrary text (including former example "device"). Do not invent pairs.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes: empty dictionary → zero matches for any text.
func TestPortugueseReplaceRule_Rule(t *testing.T) {
	// Java: new PortugueseReplaceRule(messages, language)
	rule := NewPortugueseReplaceRule(nil)

	// Former Java example (commented out): device → dispositivo is NOT live data.
	// Official replace.txt has no pairs → no match.
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("device"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("um device qualquer"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nada a substituir aqui."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("dispositivo"))))
}

// Field twin of PortugueseReplaceRule constructor + overrides.
func TestPortugueseReplaceRule_Metadata(t *testing.T) {
	rule := NewPortugueseReplaceRule(nil)

	require.Equal(t, PortugueseReplaceID, rule.GetID())
	require.Equal(t, "PT_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "Palavras estrangeiras facilmente confundidas em Português", rule.GetDescription())
	require.Equal(t, "Estrangeirismo", rule.GetShort())
	// MessageFn with empty token/replacements (static surface for metadata probes)
	require.Equal(t,
		"'' é um estrangeirismo. Em Português é mais comum usar: .",
		rule.GetMessage())

	require.Equal(t, "pt", rule.GetLocaleCode())
	require.False(t, rule.CaseSensitive)
	require.True(t, rule.CheckLemmas)
	require.True(t, rule.SubRuleSpecificIDs)
	require.Equal(t, PortugueseReplaceURL, rule.GetURL())
	require.Equal(t, "https://pt.wikipedia.org/wiki/Estrangeirismo", rule.GetURL())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSLocaleViolation, rule.GetLocQualityIssueType())

	// Official dictionary loaded: comments-only → empty wrong-words map
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, 0, len(ww), "official /pt/replace.txt has no active pairs")
	// Do not invent: device must not be in the map
	_, hasDevice := ww["device"]
	require.False(t, hasDevice)

	// Category name honors supplied ResourceBundle map
	ptMsgs := map[string]string{
		"category_style": "Estilo",
	}
	rulePT := NewPortugueseReplaceRule(ptMsgs)
	require.Equal(t, "Estilo", rulePT.GetCategory().GetName())
	// Hardcoded PT description/short do not depend on messages
	require.Equal(t, "Palavras estrangeiras facilmente confundidas em Português", rulePT.GetDescription())
	require.Equal(t, "Estrangeirismo", rulePT.GetShort())

	// Java addExamplePair is commented out → zero examples
	require.Len(t, rule.GetIncorrectExamples(), 0)
	require.Len(t, rule.GetCorrectExamples(), 0)
}

// Load does not panic (comments-only resource still parses).
func TestPortugueseReplaceRule_Loads(t *testing.T) {
	rule := NewPortugueseReplaceRule(nil)
	require.NotNil(t, rule)
	require.NotNil(t, rule.GetWrongWords())
}
