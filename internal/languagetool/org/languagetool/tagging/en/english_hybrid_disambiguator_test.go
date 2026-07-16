package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

type noopDis struct{}

func (noopDis) Disambiguate(s *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return s
}

func TestEnglishHybridDisambiguator(t *testing.T) {
	d := NewEnglishHybridDisambiguator()
	d.Chunker = noopDis{}
	d.RulesDisambiguator = noopDis{}
	sent := languagetool.AnalyzePlain("hello")
	require.Equal(t, "hello", d.Disambiguate(sent).GetText())
	require.Equal(t, sent, d.PreDisambiguate(sent))
}
