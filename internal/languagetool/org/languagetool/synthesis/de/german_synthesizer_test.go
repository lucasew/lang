package de

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestGermanSynthesizer_WithManual(t *testing.T) {
	// form\tlemma\tpos
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader("Häuser\tHaus\tSUB:NOM:PLU:NEU\n"))
	require.NoError(t, err)
	s := NewGermanSynthesizer(manual)
	lemma := "Haus"
	tok := languagetool.NewAnalyzedToken("Haus", nil, &lemma)
	forms, err := s.Synthesize(tok, "SUB:NOM:PLU:NEU")
	require.NoError(t, err)
	require.Contains(t, forms, "Häuser")
}
