package de

// Twin of GermanHelperTest.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanHelper_GetDeterminerNumber(t *testing.T) {
	require.Equal(t, "SIN", GetDeterminerNumber("ART:DEF:DAT:SIN:FEM"))
}

func TestGermanHelper_GetDeterminerDefiniteness(t *testing.T) {
	require.Equal(t, "DEF", GetDeterminerDefiniteness("ART:DEF:DAT:SIN:FEM"))
}

func TestGermanHelper_GetDeterminerCase(t *testing.T) {
	require.Equal(t, "DAT", GetDeterminerCase("ART:DEF:DAT:SIN:FEM"))
}

func TestGermanHelper_GetDeterminerGender(t *testing.T) {
	require.Equal(t, "", GetDeterminerGender(""))
	require.Equal(t, "FEM", GetDeterminerGender("ART:DEF:DAT:SIN:FEM"))
}

func TestGermanHelper_HasReadingOfType(t *testing.T) {
	tag := "ART:DEF:DAT:SIN:FEM"
	tok := languagetool.NewAnalyzedTokenStr("der", tag, "", false, true)
	readings := languagetool.NewAnalyzedTokenReadings(tok)
	require.True(t, HasReadingOfType(readings, POSDeterminer))
	require.False(t, HasReadingOfType(readings, POSNomen))
}

func TestGermanHelper_PosTypeFromAnalyzedGermanToken(t *testing.T) {
	// Java AnalyzedGermanToken type resolution
	sub := "SUB:NOM:SIN:NEU"
	require.Equal(t, POSNomen, posTypeFromAnalyzedGermanToken(&sub))
	eig := "EIG:NOM:SIN:MAS"
	require.Equal(t, POSProperNoun, posTypeFromAnalyzedGermanToken(&eig))
	// PA2 overrides earlier type when present as a part
	pa := "VER:PA2:SFT"
	require.Equal(t, POSPartizip, posTypeFromAnalyzedGermanToken(&pa))
	// short tags (<3 parts) → no type
	adv := "ADV"
	require.Equal(t, POSOther, posTypeFromAnalyzedGermanToken(&adv))
	// ADJ not PA
	adj := "ADJ:NOM:SIN:NEU:GRU:SOL"
	require.Equal(t, POSAdjective, posTypeFromAnalyzedGermanToken(&adj))
}

func TestGermanHelper_GetNounFields(t *testing.T) {
	require.Equal(t, "AKK", GetNounCase("SUB:AKK:SIN:NEU"))
	require.Equal(t, "SIN", GetNounNumber("SUB:AKK:SIN:NEU"))
	require.Equal(t, "NEU", GetNounGender("SUB:AKK:SIN:NEU"))
}
