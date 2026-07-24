package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestNonSignificantVerbsRule_DefaultMinPerMill(t *testing.T) {
	rule := NewNonSignificantVerbsRule(nil)
	require.Equal(t, nonSignificantDefaultMinPerMill, rule.MinPercent)
	require.Equal(t, 8, NewNonSignificantVerbsRuleWithDefaultLimit(nil).MinPercent)
}

func TestNonSignificantVerbsRule_MachenAngstException(t *testing.T) {
	rule := NewNonSignificantVerbsRule(nil)
	machte := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("machte", "VER:3:SIN:PRT:SFT", "machen"),
		atrWithPOS("einen", "ART:IND:AKK:SIN:MAS", "ein"),
		atrWithPOS("Kuchen", "SUB:AKK:SIN:MAS", "Kuchen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(machte)))

	angst := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "PRO:DEM:NOM:SIN:NEU", "das"),
		atrWithPOS("macht", "VER:3:SIN:PRS:SFT", "machen"),
		atrWithPOS("mir", "PRO:PER:DAT:SIN:MAS", "ich"),
		atrWithPOS("Angst", "SUB:AKK:SIN:FEM", "Angst"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(angst)))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Er machte einen Kuchen."))))
}

func TestIsUnknownWordNS_JavaRegex(t *testing.T) {
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

func TestNonSignificantVerbsRule_GetLimitMessage(t *testing.T) {
	r := NewNonSignificantVerbsRule(nil)
	require.Contains(t, r.getLimitMessage(0, 0), "wenig Aussagekraft")
	require.Contains(t, r.getLimitMessage(8, 12.4), "8‰")
	require.Contains(t, r.getLimitMessage(8, 12.4), "12‰") // 12.4+0.5 → 12
}
