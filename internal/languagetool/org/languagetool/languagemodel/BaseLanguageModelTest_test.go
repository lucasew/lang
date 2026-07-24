package languagemodel

// Twin of BaseLanguageModelTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBaseLanguageModel_PseudoProbability(t *testing.T) {
	m := NewMapLanguageModel()
	m.Total = 100
	m.Add([]string{"There"}, 10)
	m.Add([]string{"are"}, 5)
	m.Add([]string{"There", "are"}, 4)
	p := m.GetPseudoProbability([]string{"no", "data", "here"})
	require.Greater(t, p.GetProb(), 0.0) // add-1 smoothing
	p2 := m.GetPseudoProbability([]string{"There", "are"})
	require.Greater(t, p2.GetProb(), 0.0)
}

func TestBaseLanguageModel_PseudoProbabilityFail1(t *testing.T) {
	m := NewMapLanguageModel()
	require.Panics(t, func() { _ = m.GetPseudoProbability(nil) })
	require.Panics(t, func() { _ = m.GetPseudoProbability([]string{}) })
}
