package bigdata

// Twin of LanguageModelSanityTest — MapLanguageModel inject (full Lucene index deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/languagemodel"
	"github.com/stretchr/testify/require"
)

// Port of LanguageModelSanityTest (no @Test) — probability monotonicity / coverage soft.
func TestLanguageModelSanity_NoTests(t *testing.T) {
	m := languagemodel.NewMapLanguageModel()
	m.Add([]string{"the"}, 100)
	m.Add([]string{"cat"}, 50)
	m.Add([]string{"sat"}, 40)
	m.Add([]string{"the", "cat"}, 30)
	m.Add([]string{"cat", "sat"}, 20)
	m.Add([]string{"the", "cat", "sat"}, 10)
	m.Total = 1000

	// unigram counts
	require.Equal(t, int64(100), m.GetCountToken("the"))
	require.Equal(t, int64(50), m.GetCountToken("cat"))

	// bigram / trigram
	require.Equal(t, int64(30), m.GetCount([]string{"the", "cat"}))
	require.Equal(t, int64(10), m.GetCount([]string{"the", "cat", "sat"}))

	// pseudo-prob defined and in (0,1]
	p := m.GetPseudoProbability([]string{"the", "cat", "sat"})
	require.Greater(t, p.GetProb(), 0.0)
	require.LessOrEqual(t, p.GetProb(), 1.0)

	// unknown context still returns a probability (smoothing soft)
	p2 := m.GetPseudoProbability([]string{"the", "xyzzy", "sat"})
	require.GreaterOrEqual(t, p2.GetProb(), 0.0)

	// frequency index helper still available
	require.False(t, IsRealPOSTag("_START_"))
	require.NoError(t, m.Close())
}
