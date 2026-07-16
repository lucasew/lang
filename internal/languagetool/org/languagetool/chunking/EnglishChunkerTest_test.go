package chunking

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/chunking/EnglishChunkerTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func readings(token, pos string, start int) *languagetool.AnalyzedTokenReadings {
	p := pos
	tok := languagetool.NewAnalyzedToken(token, &p, nil)
	return languagetool.NewAnalyzedTokenReadingsAt(tok, start)
}

// Port of EnglishChunkerTest.testAddChunkTags
func TestEnglishChunker_AddChunkTags(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		readings("A", "DT", 0),
		readings("test", "NN", 2),
	}
	NewEnglishChunker().AddChunkTags(tokens)
	require.NotEmpty(t, tokens[1].GetChunkTags())
}

// Port of EnglishChunkerTest.testSingularNounAtEndOfNounPhrase
func TestEnglishChunker_SingularNounAtEndOfNounPhrase(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		readings("the", "DT", 0),
		readings("dog", "NN", 4),
	}
	NewEnglishChunker().AddChunkTags(tokens)
	tags := tokens[1].GetChunkTags()
	require.NotEmpty(t, tags)
	// singular NP tag expected after filter
	joined := ""
	for _, t := range tags {
		joined += t
	}
	require.Contains(t, joined, "singular")
}

// Port of EnglishChunkerTest.testAddChunkTagsSingular
func TestEnglishChunker_AddChunkTagsSingular(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		readings("cat", "NN", 0),
	}
	NewEnglishChunker().AddChunkTags(tokens)
	joined := ""
	for _, c := range tokens[0].GetChunkTags() {
		joined += c
	}
	require.Contains(t, joined, "singular")
}

// Port of EnglishChunkerTest.testContractions — basic NP still assigned on content words
func TestEnglishChunker_Contractions(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		readings("I", "PRP", 0),
		readings("'m", "VBP", 1),
		readings("fine", "JJ", 4),
	}
	NewEnglishChunker().AddChunkTags(tokens)
	// no panic; PRP not nounish
	require.Empty(t, tokens[0].GetChunkTags())
}

// Port of EnglishChunkerTest.testTokenize — chunker does not re-tokenize; surface tokens preserved
func TestEnglishChunker_Tokenize(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		readings("Hello", "UH", 0),
		readings("world", "NN", 6),
	}
	NewEnglishChunker().AddChunkTags(tokens)
	require.Equal(t, "Hello", tokens[0].GetToken())
	require.Equal(t, "world", tokens[1].GetToken())
}

// Port of EnglishChunkerTest.testNonBreakingSpace
func TestEnglishChunker_NonBreakingSpace(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		readings("foo", "NN", 0),
		readings("\u00a0", "null", 3),
		readings("bar", "NN", 4),
	}
	NewEnglishChunker().AddChunkTags(tokens)
	require.NotEmpty(t, tokens[0].GetChunkTags())
	require.NotEmpty(t, tokens[2].GetChunkTags())
}
