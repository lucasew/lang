package de

// Twin of org.languagetool.tokenizers.de.GermanCompoundTokenizerTest
// (active @Test test(); @Ignore interactive test omitted).

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGermanCompoundTokenizer_Test(t *testing.T) {
	// Java: GermanCompoundTokenizer tokenizer = new GermanCompoundTokenizer(true);
	tokenizer := NewGermanCompoundTokenizer(true)

	// assertThat(tokenizer.tokenize("Bahnhofsuhr").toString(), is("[Bahnhofs, uhr]"));
	require.Equal(t, []string{"Bahnhofs", "uhr"}, tokenizer.Tokenize("Bahnhofsuhr"))
	// assertThat(tokenizer.tokenize("natofreundlich").toString(), is("[nato, freundlich]"));
	require.Equal(t, []string{"nato", "freundlich"}, tokenizer.Tokenize("natofreundlich"))
	// assertThat(tokenizer.tokenize("natofreundliches").toString(), is("[nato, freundliches]"));
	require.Equal(t, []string{"nato", "freundliches"}, tokenizer.Tokenize("natofreundliches"))
	// assertThat(tokenizer.tokenize("Firefox-Add-on").toString(), is("[Firefox, , Add-on]"));  // why the space?
	require.Equal(t, []string{"Firefox", "", "Add-on"}, tokenizer.Tokenize("Firefox-Add-on"))
}
