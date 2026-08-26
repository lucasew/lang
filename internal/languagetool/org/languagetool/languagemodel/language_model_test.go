package languagemodel

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapLanguageModel_ChainRule(t *testing.T) {
	m := NewMapLanguageModel()
	m.Add([]string{"the"}, 100)
	m.Add([]string{"cat"}, 50)
	m.Add([]string{"the", "cat"}, 40)
	m.Add([]string{"the", "cat", "sat"}, 10)
	m.Total = 1000

	p := m.GetPseudoProbability([]string{"the", "cat", "sat"})
	require.Greater(t, p.GetProb(), 0.0)
	require.Greater(t, p.GetCoverage(), float32(0))
}

func TestMultiLanguageModel(t *testing.T) {
	a := UniformLanguageModel(0.1, 0.5)
	b := UniformLanguageModel(0.2, 0.5)
	multi := NewMultiLanguageModel([]LanguageModel{a, b})
	p := multi.GetPseudoProbability([]string{"x"})
	require.InDelta(t, 0.3, p.GetProb(), 1e-9)
	require.InDelta(t, 0.5, float64(p.GetCoverage()), 1e-6)
	require.Panics(t, func() { NewMultiLanguageModel(nil) })
}

func TestStupidBackoff(t *testing.T) {
	m := NewMapLanguageModel()
	m.Add([]string{"a", "b"}, 5)
	m.Add([]string{"a"}, 10)
	p := m.base.GetPseudoProbabilityStupidBackoff([]string{"a", "b", "c"})
	// falls back eventually
	require.GreaterOrEqual(t, p.GetProb(), 0.0)
}
