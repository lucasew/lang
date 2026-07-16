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
