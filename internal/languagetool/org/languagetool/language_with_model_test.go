package languagetool

import (
	"os"
	"path/filepath"
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
	dir := t.TempDir()
	// Java: topIndexDir = new File(indexDir, getShortCode())
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "en"), 0o755))

	l := NewLanguageWithModel("en", "English")
	l.InitModel = func(topIndexDir string) (NgramModel, error) {
		require.Equal(t, filepath.Join(dir, "en"), topIndexDir)
		return fakeModel{}, nil
	}
	m, err := l.GetLanguageModel(dir)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Equal(t, 0.5, m.GetPseudoProbability(nil).GetProb())
	m2, err := l.GetLanguageModel(dir)
	require.NoError(t, err)
	require.Equal(t, m, m2)
	require.NoError(t, l.Close())
}

func TestLanguageWithModel_MissingDirWarnsOnce(t *testing.T) {
	l := NewLanguageWithModel("zz", "ZedZed")
	m, err := l.GetLanguageModel(t.TempDir())
	require.NoError(t, err)
	require.Nil(t, m)
	// second call does not re-init warning path for model; still nil
	m2, err := l.GetLanguageModel(t.TempDir())
	require.NoError(t, err)
	require.Nil(t, m2)
}
