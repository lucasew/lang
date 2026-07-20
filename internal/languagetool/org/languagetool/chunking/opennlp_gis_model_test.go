package chunking

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadGISModel_Chunker(t *testing.T) {
	p := DiscoverOpenNLPChunkerModel()
	if p == "" {
		t.Skip("en-chunker.bin not found under third_party/opennlp-models")
	}
	m, err := LoadGISModelFromZip(p)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Greater(t, m.NumOutcomes(), 5)
	// trivial context — eval must sum to ~1
	probs := m.Eval([]string{"w0=the", "t0=DT", "w1=dog", "t1=NN"})
	require.Len(t, probs, m.NumOutcomes())
	var sum float64
	for _, p := range probs {
		require.GreaterOrEqual(t, p, 0.0)
		sum += p
	}
	require.InDelta(t, 1.0, sum, 1e-6)
}

func TestChunkerME_Runs(t *testing.T) {
	p := DiscoverOpenNLPChunkerModel()
	if p == "" {
		t.Skip("en-chunker.bin not found")
	}
	c, err := NewChunkerME(p)
	require.NoError(t, err)
	toks := []string{"The", "quick", "brown", "fox", "jumps"}
	tags := []string{"DT", "JJ", "JJ", "NN", "VBZ"}
	chunks := c.Chunk(toks, tags)
	require.Len(t, chunks, len(toks))
	for _, ch := range chunks {
		require.NotEmpty(t, ch)
	}
	// Expect an NP somewhere (OpenNLP BIO)
	joined := ""
	for _, ch := range chunks {
		joined += ch + " "
	}
	require.Contains(t, joined, "NP")
}
