package de

// Twin of GermanTaggerTest — MapWordTagger smokes; full dict deferred.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestGermanTagger_AdjectivesFromSpellingTxt(t *testing.T) {
	wt := tagging.MapWordTagger{"schön": {tagging.NewTaggedWord("schön", "ADJ:PRD:GRU")}}
	got := NewGermanTagger(wt).Tag([]string{"schön"})
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestGermanTagger_LemmaOfForDashCompounds(t *testing.T) {
	// Without compound splitter, full form lookup works when present.
	wt := tagging.MapWordTagger{
		"Diabetes-Zentrum": {tagging.NewTaggedWord("Diabetes-Zentrum", "SUB:NOM:SIN:NEU")},
	}
	got := NewGermanTagger(wt).Tag([]string{"Diabetes-Zentrum"})
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestGermanTagger_GenderGap(t *testing.T) {
	// Surface still tags known lemmas; gender-gap forms need extended rules.
	wt := tagging.MapWordTagger{"Student": {tagging.NewTaggedWord("Student", "SUB:NOM:SIN:MAS")}}
	got := NewGermanTagger(wt).Tag([]string{"Student"})
	require.NotEmpty(t, got[0].GetReadings())
}

func TestGermanTagger_IgnoreDomain(t *testing.T) {
	// Domains remain untagged without dict entry
	got := NewGermanTagger(tagging.MapWordTagger{}).Tag([]string{"example.com"})
	require.Nil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestGermanTagger_IgnoreImperative(t *testing.T) {
	wt := tagging.MapWordTagger{"geh": {tagging.NewTaggedWord("gehen", "VER:IMP:SIN:SFT")}}
	got := NewGermanTagger(wt).Tag([]string{"geh"})
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestGermanTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Der":  {tagging.NewTaggedWord("der", "ART:DEF:NOM:SIN:MAS")},
		"Hund": {tagging.NewTaggedWord("Hund", "SUB:NOM:SIN:MAS")},
	}
	got := NewGermanTagger(wt).Tag([]string{"Der", "Hund", "xyz"})
	require.Len(t, got, 3)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.NotNil(t, got[1].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[2].GetReadings()[0].GetPOSTag())
}

func TestGermanTagger_ExtendedTagger(t *testing.T) {
	wt := tagging.MapWordTagger{"Häuser": {tagging.NewTaggedWord("Haus", "SUB:NOM:PLU:NEU")}}
	got := NewGermanTagger(wt).Tag([]string{"Häuser"})
	require.Equal(t, "SUB:NOM:PLU:NEU", *got[0].GetReadings()[0].GetPOSTag())
}

func TestGermanTagger_AfterColon(t *testing.T) {
	// After-colon casing: MapWordTagger can provide both cases.
	wt := tagging.MapWordTagger{
		"Hallo": {tagging.NewTaggedWord("Hallo", "ITJ")},
		"welt":  {tagging.NewTaggedWord("Welt", "SUB:NOM:SIN:FEM")},
	}
	got := NewGermanTagger(wt).Tag([]string{"Hallo", ":", "welt"})
	require.Len(t, got, 3)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestGermanTagger_Dictionary(t *testing.T) {
	wt := tagging.MapWordTagger{"Tisch": {tagging.NewTaggedWord("Tisch", "SUB:NOM:SIN:MAS")}}
	tagger := NewGermanTagger(wt)
	require.Equal(t, GermanDictPath, tagger.GetDictionaryPath())
	require.Len(t, tagger.TagWord("Tisch"), 1)
}
