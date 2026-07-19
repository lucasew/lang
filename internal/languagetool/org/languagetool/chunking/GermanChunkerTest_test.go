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
	// REGEXES2 may add NPS on top of B-NP/I-NP (Java additive tags).
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[1].GetChunkTags(), "I-NP")
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
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[1].GetChunkTags(), "I-NP")
	require.Contains(t, tokens[2].GetChunkTags(), "I-NP")
}

// Port of GermanChunkerTest.testTemp — bare EIG alone is not REGEXES1 (needs Herr+EIG etc.).
// Java leaves untagged tokens as O; no invent POS→BIO fallback.
func TestGermanChunker_Temp(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		deTok("Berlin", "EIG:NOM:SIN:NEU", 0),
	}
	NewGermanChunker().AddChunkTags(tokens)
	// No B-NP invent for lone EIG (REGEXES1 has no bare-EIG pattern).
	require.NotContains(t, tokens[0].GetChunkTags(), "B-NP")
	require.NotContains(t, tokens[0].GetChunkTags(), "I-NP")
}
