package languagemodel

// Twin of LanguageModelTest (Java has no @Test) — MapLanguageModel smoke.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLanguageModel_NoTests(t *testing.T) {
	m := NewMapLanguageModel()
	m.Total = 100
	m.Add([]string{"the"}, 50)
	m.Add([]string{"cat"}, 10)
	m.Add([]string{"the", "cat"}, 5)
	p := m.GetPseudoProbability([]string{"the", "cat"})
	require.Greater(t, p.GetProb(), 0.0)
	require.NoError(t, m.Close())
}
