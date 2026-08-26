package chunking

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/chunking/EnglishChunkFilterTest.java
import (
	"fmt"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of EnglishChunkFilterTest.testSingular
func TestEnglishChunkFilter_Singular(t *testing.T) {
	assertChunks(t,
		"He/B-NP owns/B-VP a/B-NP nice/I-NP house/I-NP in/X Berlin/B-NP ./.",
		"He/B-NP-singular,E-NP-singular owns/B-VP a/B-NP-singular nice/I-NP-singular house/E-NP-singular in/X Berlin/B-NP-singular,E-NP-singular ./.",
	)
}

// Port of EnglishChunkFilterTest.testPluralByAnd — Java @Ignore("fails...")
// Do not invent a green twin that forces ignored Java behavior.
func TestEnglishChunkFilter_PluralByAnd(t *testing.T) {
	t.Skip("Java @Ignore(\"fails...\") — not forced green")
	assertChunks(t,
		"He/B-NP owns/B-VP a/B-NP large/I-NP house/I-NP and/I-NP a/I-NP ship/I-NP in/X Berlin/B-NP ./.",
		"He/B-NP-singular owns/B-VP a/B-NP-plural large/I-NP-plural house/I-NP-plural and/I-NP-plural a/I-NP-plural ship/I-NP-plural in/X Berlin/B-NP-singular ./.",
	)
}

// Port of EnglishChunkFilterTest.testPluralByPluralNoun
func TestEnglishChunkFilter_PluralByPluralNoun(t *testing.T) {
	input := "I/X have/N-VP ten/B-NP books/I-NP ./."
	tokens := makeTokens(input)
	// Java: tokens.remove(3); // 'books'
	tokens = append(tokens[:3], tokens[4:]...)
	posNNS := "NNS"
	posVBZ := "VBZ"
	lemma := "book"
	readings := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("books", &posNNS, &lemma),
		languagetool.NewAnalyzedToken("books", &posVBZ, &lemma),
	}, 0)
	// Java: tokens.add(3, new ChunkTaggedToken("books", singletonList(I-NP), readings))
	books := NewChunkTaggedToken("books", []ChunkTag{NewChunkTag("I-NP")}, readings)
	tokens = append(tokens[:3], append([]ChunkTaggedToken{books}, tokens[3:]...)...)
	assertChunksTokens(t, tokens, "I/X have/N-VP ten/B-NP-plural books/E-NP-plural ./.")
}

// assertChunks ports EnglishChunkFilterTest.assertChunks(String, String)
func assertChunks(t *testing.T, input, expected string) {
	t.Helper()
	assertChunksTokens(t, makeTokens(input), expected)
}

// assertChunksTokens ports EnglishChunkFilterTest.assertChunks(List, String)
// Java: StringUtils.join(result, " ") via ChunkTaggedToken.toString()
func assertChunksTokens(t *testing.T, tokens []ChunkTaggedToken, expected string) {
	t.Helper()
	filter := NewEnglishChunkFilter()
	result := filter.Filter(tokens)
	parts := make([]string, len(result))
	for i, tok := range result {
		parts[i] = tok.String()
	}
	require.Equal(t, expected, strings.Join(parts, " "))
}

// makeTokens ports EnglishChunkFilterTest.makeTokens
func makeTokens(tokensAsString string) []ChunkTaggedToken {
	var result []ChunkTaggedToken
	for _, token := range strings.Split(tokensAsString, " ") {
		parts := strings.Split(token, "/")
		if len(parts) != 2 {
			panic(fmt.Sprintf("Invalid token, form 'x/y' required: %s", token))
		}
		chunkTag := NewChunkTag(parts[1])
		result = append(result, NewChunkTaggedToken(parts[0], []ChunkTag{chunkTag}, nil))
	}
	return result
}
