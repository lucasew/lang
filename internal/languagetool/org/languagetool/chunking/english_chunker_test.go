package chunking

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnglishChunker(t *testing.T) {
	// Spaced tokens like Java createReadingsList so OpenNLP position map works.
	tokens := createReadingsList("The dogs run")
	NewEnglishChunker().AddChunkTags(tokens)
	// dogs (index 2) should get NP-plural chunk tags
	require.NotEmpty(t, tokens[2].GetChunkTags())
	joined := ""
	for _, c := range tokens[2].GetChunkTags() {
		joined += c
	}
	require.Contains(t, joined, "NP")
}
