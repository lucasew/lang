package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishSynthesizerDeterminers(t *testing.T) {
	s := NewEnglishSynthesizer(nil)
	tok := languagetool.NewAnalyzedToken("apple", nil, strp("apple"))
	got, err := s.Synthesize(tok, AddIndDeterminer)
	require.NoError(t, err)
	require.Equal(t, []string{"an apple"}, got)
	got, err = s.Synthesize(tok, AddDeterminer)
	require.NoError(t, err)
	require.Contains(t, got, "an apple")
	require.Contains(t, got, "the apple")
}

func strp(s string) *string { return &s }
