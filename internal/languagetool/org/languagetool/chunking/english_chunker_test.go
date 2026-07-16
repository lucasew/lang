package chunking

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishChunker(t *testing.T) {
	nn := "NN"
	nns := "NNS"
	det := "DT"
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("The", &det, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("dogs", &nns, nil), 4),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("run", &nn, nil), 9),
	}
	NewEnglishChunker().AddChunkTags(tokens)
	// dogs should get NP-plural chunk tags
	require.NotEmpty(t, tokens[1].GetChunkTags())
}
