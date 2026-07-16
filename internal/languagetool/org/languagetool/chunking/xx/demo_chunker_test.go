package xx

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDemoChunker(t *testing.T) {
	c := NewDemoChunker()
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("hello", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("chunkbar", nil, nil)),
	}
	c.AddChunkTags(toks)
	require.Empty(t, toks[0].GetChunkTags())
	require.Equal(t, []string{"B-NP-singular"}, toks[1].GetChunkTags())
}
