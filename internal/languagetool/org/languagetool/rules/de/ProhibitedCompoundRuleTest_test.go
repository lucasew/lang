package de

// Twin of ProhibitedCompoundRuleTest (surface preferred-direction fragments).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestProhibitedCompoundRule_Rule(t *testing.T) {
	rule := NewProhibitedCompoundRule(nil)
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	require.Equal(t, 1, matchN("Da steht eine Lehrzeile zu viel."))
	require.Equal(t, 0, matchN("Eine Leerzeile einfügen."))
	require.Equal(t, 1, matchN("Das ist ein Mitauto."))
	require.Equal(t, 1, matchN("Er ist Uhrberliner."))
	require.Equal(t, 1, matchN("Hier leben die Uhreinwohner."))
	require.Equal(t, 1, matchN("Eine Lehr-Zeile einfügen."))
	require.Equal(t, 0, matchN("Das ist Herr Mitauto."))
}

func TestProhibitedCompoundRule_RemoveHyphensAndAdaptCase(t *testing.T) {
	require.Equal(t, "", removeHyphensAndAdaptCase("Marathonläuse"))
	require.Equal(t, "Marathonläuse", removeHyphensAndAdaptCase("Marathon-Läuse"))
	require.Equal(t, "Marathonläusetest", removeHyphensAndAdaptCase("Marathon-Läuse-Test"))
	require.Equal(t, "", removeHyphensAndAdaptCase("S-Bahn"))
}
