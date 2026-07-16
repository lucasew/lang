package chunking

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/chunking/GermanChunkerTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func deTok(token, pos string, start int) *languagetool.AnalyzedTokenReadings {
	p := pos
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(token, &p, nil), start)
}

// Port of GermanChunkerTest.testChunking
func TestGermanChunker_Chunking(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		deTok("Der", "ART:DEF:NOM:SIN:MAS", 0),
		deTok("Hund", "SUB:NOM:SIN:MAS", 4),
		deTok("bellt", "VER:3:SIN:PRÄ:SFT", 9),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Equal(t, []string{"B-NP"}, tokens[0].GetChunkTags())
	require.Equal(t, []string{"I-NP"}, tokens[1].GetChunkTags())
	require.Empty(t, tokens[2].GetChunkTags())
}

// Port of GermanChunkerTest.testOpenNLPLikeChunking
func TestGermanChunker_OpenNLPLikeChunking(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		deTok("Ein", "ART:IND:NOM:SIN:NEU", 0),
		deTok("schönes", "ADJ:NOM:SIN:NEU:GRU:SOL", 4),
		deTok("Haus", "SUB:NOM:SIN:NEU", 12),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Equal(t, []string{"B-NP"}, tokens[0].GetChunkTags())
	require.Equal(t, []string{"I-NP"}, tokens[1].GetChunkTags())
	require.Equal(t, []string{"I-NP"}, tokens[2].GetChunkTags())
}

// Port of GermanChunkerTest.testTemp — short smoke
func TestGermanChunker_Temp(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		deTok("Berlin", "EIG:NOM:SIN:NEU", 0),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Equal(t, []string{"B-NP"}, tokens[0].GetChunkTags())
}
