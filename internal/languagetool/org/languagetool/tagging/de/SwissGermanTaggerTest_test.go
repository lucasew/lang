package de

// Twin of SwissGermanTaggerTest — MapWordTagger ss↔ß mapping
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestSwissGermanTagger_Tagger(t *testing.T) {
	// DE dict has ß spellings; Swiss uses ss. Lookup uses ignoreCase=false so ß forms
	// must match exact surface (as in german.dict).
	wt := tagging.MapWordTagger{
		"groß":     {tagging.NewTaggedWord("groß", "ADJ:PRD:GRU")},
		"Anmaßung": {tagging.NewTaggedWord("Anmaßung", "SUB:NOM:SIN:FEM")},
		"die":      {tagging.NewTaggedWord("die", "ART:DEF:NOM:SIN:FEM")},
		"Auto":     {tagging.NewTaggedWord("Auto", "SUB:NOM:SIN:NEU")},
	}
	swiss := NewSwissGermanTagger(wt)
	german := NewGermanTagger(wt)

	aToken := swiss.Lookup("gross")
	require.NotNil(t, aToken)
	require.NotNil(t, aToken.GetReadings()[0].GetPOSTag())
	require.Equal(t, "gross", aToken.GetReadings()[0].GetToken())
	require.Equal(t, "groß", *aToken.GetReadings()[0].GetLemma())
	require.Equal(t, "ADJ:PRD:GRU", *aToken.GetReadings()[0].GetPOSTag())

	aToken2 := swiss.Lookup("Anmassung")
	require.NotNil(t, aToken2.GetReadings()[0].GetPOSTag())
	require.Equal(t, "Anmaßung", *aToken2.GetReadings()[0].GetLemma())
	require.Equal(t, "Anmassung", aToken2.GetReadings()[0].GetToken())

	// same as German for words without ss
	require.Equal(t,
		german.Lookup("die").GetReadings()[0].GetToken(),
		swiss.Lookup("die").GetReadings()[0].GetToken())
	require.Equal(t,
		*german.Lookup("Auto").GetReadings()[0].GetLemma(),
		*swiss.Lookup("Auto").GetReadings()[0].GetLemma())
}
