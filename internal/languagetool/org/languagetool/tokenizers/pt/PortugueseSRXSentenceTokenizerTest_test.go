package pt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Ports key cases from Java PortugueseSRXSentenceTokenizerTest (segment.srx).
func TestPortugueseSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewPortugueseSRXSentenceTokenizer()
	require.NotNil(t, tok)

	// Two sentences (space after period, as in Java tests).
	got := tok.Tokenize("Cola o teu próprio texto aqui. Ou verifica este texto.")
	require.Len(t, got, 2)
	require.Contains(t, got[0], "aqui.")

	// Abbreviations must not split (Java: Sr., Dra., etc.).
	for _, text := range []string{
		"O Sr. João foi ao mercado.",
		"A Dra. Ana é especialista em pediatria.",
		"Comprei frutas, legumes, etc. no supermercado.",
		// Portuguese abbreviated ordinal "12o." (segment.srx \\d+o) — one sentence.
		"O premiado é o 12o. da lista.",
		"O 1.º lugar foi do Brasil.",
	} {
		parts := tok.Tokenize(text)
		require.Lenf(t, parts, 1, "text=%q parts=%#v", text, parts)
		require.Equal(t, text, strings.Join(parts, ""))
	}
}
