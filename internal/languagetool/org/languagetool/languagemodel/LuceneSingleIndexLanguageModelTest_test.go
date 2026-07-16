package languagemodel

// Twin of LuceneSingleIndexLanguageModelTest — Lucene deferred; MapLanguageModel stand-in.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLuceneSingleIndexLanguageModel_LanguageModel(t *testing.T) {
	m := NewMapLanguageModel()
	m.Total = 1000
	m.Add([]string{"the"}, 100)
	m.Add([]string{"cat"}, 10)
	m.Add([]string{"the", "cat"}, 5)
	p := m.GetPseudoProbability([]string{"the", "cat"})
	require.Greater(t, p.GetProb(), 0.0)
	require.Greater(t, p.GetCoverage(), float32(0))
}
