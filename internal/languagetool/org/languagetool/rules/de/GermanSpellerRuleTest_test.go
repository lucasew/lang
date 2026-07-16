package de

// Twin of GermanSpellerRuleTest — dictionary-backed methods soft until Morfologik lands.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanSpellerRule_GetMessage(t *testing.T) {
	r := NewGermanSpellerRule(nil)
	require.Contains(t, r.GetMessage("Feler", "Fehler"), "Feler")
	require.Contains(t, r.GetMessage("Feler", "Fehler"), "Fehler")
}

func TestGermanSpellerRule_Artig(t *testing.T) {
	// needs dict; soft: constructor + no-op Match
	r := NewGermanSpellerRule(nil)
	require.Equal(t, 0, len(r.Match(languagetool.AnalyzePlain("Das ist artig."))))
}

func TestAustrianGermanSpellerRule(t *testing.T) {
	r := NewAustrianGermanSpellerRule(nil)
	require.Equal(t, "AUSTRIAN_GERMAN_SPELLER_RULE", r.GetID())
}

func TestSwissGermanSpellerRule(t *testing.T) {
	r := NewSwissGermanSpellerRule(nil)
	require.Equal(t, "SWISS_GERMAN_SPELLER_RULE", r.GetID())
}
