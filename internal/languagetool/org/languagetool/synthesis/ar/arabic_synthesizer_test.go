package ar

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestArabicSynthesizer(t *testing.T) {
	man, err := synthesis.NewManualSynthesizer(strings.NewReader("كتب\tكتب\tNxx\n"))
	require.NoError(t, err)
	s := NewArabicSynthesizer(man)
	lemma := "كتب"
	tok := languagetool.NewAnalyzedToken("كتب", nil, &lemma)
	forms, err := s.Synthesize(tok, "Nxx")
	require.NoError(t, err)
	require.Equal(t, []string{"كتب"}, forms)
	require.Equal(t, ArabicSynthDict, s.ResourceFileName)
}
