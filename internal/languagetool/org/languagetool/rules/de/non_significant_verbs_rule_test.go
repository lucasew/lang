package de

// Unit helpers for NonSignificantVerbsRule pure parity (messages, defaults, isUnknownWord regex).
// Full Java testRule twin lives in NonSignificantVerbsRuleTest_test.go (real de-DE analysis).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestNonSignificantVerbsRule_DefaultMinPerMill(t *testing.T) {
	rule := NewNonSignificantVerbsRule(nil)
	require.Equal(t, nonSignificantDefaultMinPerMill, rule.MinPercent)
	require.Equal(t, 8, NewNonSignificantVerbsRuleWithDefaultLimit(nil).MinPercent)
	require.Equal(t, 0, NewNonSignificantVerbsRuleWithMinPercent(nil, 0).MinPercent)
	require.Equal(t, "NON_SIGNIFICANT_VERB_DE", rule.GetID())
	require.Equal(t, float64(1000), rule.Denominator)
	require.True(t, rule.ExcludeDirectSpeech)
	require.True(t, rule.IsDefaultOff())
}

func TestNonSignificantVerbsRule_GetLimitMessage(t *testing.T) {
	r := NewNonSignificantVerbsRule(nil)
	require.Equal(t,
		"Dieses Verb hat wenig Aussagekraft. Verwenden Sie wenn möglich ein anderes oder formulieren Sie den Satz um.",
		r.getLimitMessage(0, 0))
	// Java: "Mehr als " + limit + "‰ … {" + ((int)(percent+0.5d)) + "‰} …"
	msg := r.getLimitMessage(8, 12.4)
	require.Contains(t, msg, "Mehr als 8‰")
	require.Contains(t, msg, "12‰") // 12.4+0.5 → 12
	require.Contains(t, msg, "wenig aussagekräftige Verben")
}

func TestIsUnknownWordNS_JavaRegex(t *testing.T) {
	// Pure isUnknownWord: isPosTagUnknown && length()>2 && matches Java char class.
	// AnalyzePlain untagged surfaces — not a substitute for testRule.
	plain := languagetool.AnalyzePlain("Abcdef")
	var content *languagetool.AnalyzedTokenReadings
	for _, tok := range plain.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "Abcdef" {
			content = tok
			break
		}
	}
	require.NotNil(t, content)
	require.True(t, content.IsPosTagUnknown())
	require.True(t, isUnknownWordNS(content))

	// Java character class excludes é
	bad := atrWithPOS("caféxx", "", "")
	if bad.IsPosTagUnknown() {
		require.False(t, isUnknownWordNS(bad), "é not in Java [A-Za-zÄÖÜäöüß] class")
	}
	// length <= 2
	short := languagetool.AnalyzePlain("Ab")
	for _, tok := range short.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "Ab" {
			require.False(t, isUnknownWordNS(tok))
		}
	}
}

// TestNonSignificantVerbsRule_MachenAngstException exercises isException machen branch
// with real GermanTagger lemmas (not AnalyzePlain invent-soft).
func TestNonSignificantVerbsRule_MachenAngstException(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de-DE")
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required")
	tagger := DiscoveredGermanTagger()
	require.NotNil(t, tagger)
	rule0 := NewNonSignificantVerbsRuleWithMinPercent(nil, 0) // show-all to isolate exception

	// "Er machte einen Kuchen." — machen, no Angst/Weg… → hit (example wrong side)
	kuchen := analyzeWithGermanTagger(tagger, "Er machte einen Kuchen.")
	var machte *languagetool.AnalyzedTokenReadings
	for _, tok := range kuchen.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "machte" {
			machte = tok
			break
		}
	}
	require.NotNil(t, machte)
	require.True(t, machte.HasAnyLemma("machen"))
	require.Equal(t, 1, len(rule0.Match(kuchen)))

	// "Das macht mir Angst." — machen + Angst exception → 0
	angst := analyzeWithGermanTagger(tagger, "Das macht mir Angst.")
	require.Equal(t, 0, len(rule0.Match(angst)))
}
