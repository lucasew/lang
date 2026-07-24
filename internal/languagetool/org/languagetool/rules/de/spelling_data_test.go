package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadSpellingDataBoth(t *testing.T) {
	content := `# c
Schluß;Schluss
alt;neu
`
	d, err := LoadSpellingDataBoth(content, "test.csv", nil)
	require.NoError(t, err)
	v, ok := d.Lookup("Schluß")
	require.True(t, ok)
	require.Equal(t, "Schluss", v)
	require.Equal(t, "neu", d.Map["alt"])
	// sentence start capitalizes lowercase pairs
	require.Equal(t, "Neu", d.SentenceStartMap["Alt"])

	_, err = LoadSpellingDataBoth("same;same\n", "x", nil)
	require.Error(t, err)
}

// Port of SpellingData ß→ss expansion via synthesizeForPosTags (Java SpellingData L68–75).
func TestLoadSpellingDataBoth_ExpandForms(t *testing.T) {
	content := "Rußland;Russland\n"
	expand := func(old string) []string {
		if old == "Rußland" {
			// Schlüsse has "ss" — Java skips (new-spelling noise from old lemma Schluß).
			return []string{"Rußlands", "Schlüsse"}
		}
		return nil
	}
	d, err := LoadSpellingDataBoth(content, "t.csv", expand)
	require.NoError(t, err)
	require.Equal(t, "Russland", d.Map["Rußland"])
	require.Equal(t, "Russlands", d.Map["Rußlands"])
	_, ok := d.Map["Schlüsse"]
	require.False(t, ok)
}
