package chunking

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRussianChunker(t *testing.T) {
	nn := "NN:nom:m"
	v := "V:ipf"
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("кот", &nn, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("спит", &v, nil), 4),
	}
	NewRussianChunker().AddChunkTags(tokens)
	require.Equal(t, []string{"B-NP"}, tokens[0].GetChunkTags())
	require.Equal(t, []string{"B-VP"}, tokens[1].GetChunkTags())
}
