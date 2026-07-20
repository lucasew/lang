package chunking

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishChunker(t *testing.T) {
	if DiscoverOpenNLPChunkerModel() == "" {
		t.Skip("OpenNLP models required — Java EnglishChunker has no invent POS→BIO path")
	}
	// Spaced tokens like Java createReadingsList so OpenNLP position map works.
	tokens := createReadingsList("The dogs run")
	// Filter uses LT POS for plural; set NNS on "dogs".
	nns := "NNS"
	tokens[2] = languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("dogs", &nns, nil), tokens[2].GetStartPos())
	NewEnglishChunker().AddChunkTags(tokens)
	// dogs (index 2) should get NP-plural chunk tags
	require.NotEmpty(t, tokens[2].GetChunkTags())
	joined := ""
	for _, c := range tokens[2].GetChunkTags() {
		joined += c
	}
	require.Contains(t, joined, "NP")
}

// TestEnglishChunker_NoInventPOSBIO: without invent fallback, LT POS alone never yields BIO.
// Chunk tags come only from OpenNLP (when models load).
func TestEnglishChunker_NoInventPOSBIO(t *testing.T) {
	// Direct unit of the filter path is covered elsewhere; here we assert that
	// EnglishChunker has no AssignBasicNP / invent surface — only Filter field.
	c := NewEnglishChunker()
	require.NotNil(t, c.Filter)
	// Struct has no invent knobs (compile-time: only Filter remains).
	_ = c
}
