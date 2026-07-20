package synthesis

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func findFile(t *testing.T, rel string) string {
	t.Helper()
	dir, _ := os.Getwd()
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, rel)
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		dir = filepath.Dir(dir)
	}
	return ""
}

func TestSoros_RealEnSor(t *testing.T) {
	p := findFile(t, "inspiration/languagetool/languagetool-language-modules/en/src/main/resources/org/languagetool/resource/en/en.sor")
	if p == "" {
		t.Skip("en.sor missing")
	}
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	s := NewSoros(string(b), "en")
	// Java EnglishSynthesizerTest spell-number expectations
	require.Equal(t, "twelve", s.Run("12"))
	require.Equal(t, "zero", s.Run("0"))
	require.Equal(t, "one", s.Run("1"))
	require.Equal(t, "one thousand two hundred forty-three", s.Run("1243"))
	require.Equal(t, "one hundred", s.Run("100"))
	require.Equal(t, "twenty-one", s.Run("21"))
}

func TestSoros_RealRoman(t *testing.T) {
	p := findFile(t, "inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/Roman.sor")
	if p == "" {
		t.Skip("Roman.sor missing")
	}
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	s := NewSoros(string(b), "Roman")
	require.Equal(t, "XII", s.Run("12"))
	require.Equal(t, "IV", s.Run("4"))
	require.Equal(t, "IX", s.Run("9"))
}
