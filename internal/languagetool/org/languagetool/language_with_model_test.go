package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeProb struct{ p float64 }

func (f fakeProb) GetProb() float64     { return f.p }
func (f fakeProb) GetCoverage() float32 { return 1 }

type fakeModel struct{}

func (fakeModel) GetPseudoProbability(context []string) PseudoProbability {
	return fakeProb{0.5}
}

func TestLanguageWithModel(t *testing.T) {
	l := NewLanguageWithModel("en", "English")
	l.InitModel = func(indexDir string) (NgramModel, error) {
		return fakeModel{}, nil
	}
	m, err := l.GetLanguageModel("/tmp")
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Equal(t, 0.5, m.GetPseudoProbability(nil).GetProb())
	m2, err := l.GetLanguageModel("/tmp")
	require.NoError(t, err)
	require.Equal(t, m, m2)
	require.NoError(t, l.Close())
}
