package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUkrainianHybridDisambiguator(t *testing.T) {
	d := NewUkrainianHybridDisambiguator()
	require.Nil(t, d.Disambiguate(nil))
	s := languagetool.NewAnalyzedSentence(nil)
	require.NotNil(t, d.Disambiguate(s))
	require.NotNil(t, NewSimpleDisambiguator().Disambiguate(s))
	d2 := NewUkrainianHybridDisambiguatorWith(NewUkrainianMultiwordChunker(nil), NewSimpleDisambiguator())
	require.NotNil(t, d2.Disambiguate(s))
}
