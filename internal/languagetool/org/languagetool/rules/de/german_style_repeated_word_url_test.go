package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanStyleRepeatedWordURL(t *testing.T) {
	// single lemma → OpenThesaurus + lemma
	lemma := "gehen"
	tok := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("gehe", strPtr("VER:1:SIN:PRÄ:NON:NEB"), &lemma),
	)
	require.Equal(t, "https://www.openthesaurus.de/synonyme/gehen", germanStyleRepeatedWordURL(tok))

	// multiple lemmas → surface token
	l2 := "Haus"
	l3 := "hausen"
	tok2 := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("Haus", strPtr("SUB:NOM:SIN:NEU"), &l2),
	)
	tok2.AddReading(languagetool.NewAnalyzedToken("Haus", strPtr("VER:INF:NON"), &l3), "")
	require.Equal(t, "https://www.openthesaurus.de/synonyme/Haus", germanStyleRepeatedWordURL(tok2))

	// no lemma → surface
	tok3 := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("xyz", nil, nil),
	)
	require.Equal(t, "https://www.openthesaurus.de/synonyme/xyz", germanStyleRepeatedWordURL(tok3))

	require.Equal(t, "", germanStyleRepeatedWordURL(nil))
}

func TestGermanStyleRepeatedWordRule_SetURLWired(t *testing.T) {
	r := NewGermanStyleRepeatedWordRule(nil)
	require.NotNil(t, r.SetURL)
	// two sentences with repeated content word (needs POS or unknown word path)
	// Force via Analyze with tags is heavy; just assert SetURL is non-nil and works.
	lemma := "laufen"
	tok := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("laufe", strPtr("VER:1:SIN:PRÄ:NON:NEB"), &lemma),
	)
	require.Contains(t, r.SetURL(tok), "openthesaurus.de/synonyme/laufen")
}
