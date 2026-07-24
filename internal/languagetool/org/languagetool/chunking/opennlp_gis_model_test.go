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

// Java OpenNLP 1.9.4 ChunkerME ground truth on same en-*.bin models:
// I'll be there → I/PRP/B-NP, 'll/MD/B-VP, be/VB/I-VP, there/RB/I-VP
func TestChunkerME_IllBeThere_JavaParity(t *testing.T) {
	p := DiscoverOpenNLPChunkerModel()
	if p == "" {
		t.Skip("en-chunker.bin not found")
	}
	c, err := NewChunkerME(p)
	require.NoError(t, err)
	toks := []string{"I", "'ll", "be", "there"}
	tags := []string{"PRP", "MD", "VB", "RB"}
	chunks := c.Chunk(toks, tags)
	require.Equal(t, []string{"B-NP", "B-VP", "I-VP", "I-VP"}, chunks)
}

// DefaultChunkerContext must emit OpenNLP's p_2 quirk (no '=' when not bos).
func TestDefaultChunkerContext_P2Quirk(t *testing.T) {
	preds := []string{"B-NP", "B-VP", "I-VP"}
	toks := []string{"I", "'ll", "be", "there"}
	tags := []string{"PRP", "MD", "VB", "RB"}
	ctx := DefaultChunkerContext(3, toks, tags, preds)
	// At i=3, p_2 should be "p_2"+"B-VP" = "p_2B-VP" (not "p_2=B-VP")
	found := false
	for _, f := range ctx {
		if f == "p_2B-VP" {
			found = true
		}
		require.NotEqual(t, "p_2=B-VP", f)
	}
	require.True(t, found, "missing p_2 quirk feature p_2B-VP in %v", ctx)
}
