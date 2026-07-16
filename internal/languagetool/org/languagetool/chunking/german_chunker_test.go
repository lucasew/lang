package chunking

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanChunker(t *testing.T) {
	art := "ART:DEF:NOM:SIN:MAS"
	sub := "SUB:NOM:SIN:MAS"
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Der", &art, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Hund", &sub, nil), 4),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Equal(t, []string{"B-NP"}, tokens[0].GetChunkTags())
	require.Equal(t, []string{"I-NP"}, tokens[1].GetChunkTags())
}
