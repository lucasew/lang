package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfusionSetLoader_AlphabetOrder(t *testing.T) {
	loader := NewConfusionSetLoader(nil)
	_, err := loader.LoadConfusionPairs(strings.NewReader("zebra; apple; 1\n"))
	require.Error(t, err)
}

func TestConfusionSetLoader_WordDefsHook(t *testing.T) {
	loader := NewConfusionSetLoader(func(word string) *string {
		if word == "a" {
			d := "letter a"
			return &d
		}
		return nil
	})
	m, err := loader.LoadConfusionPairs(strings.NewReader("a; b; 3\n"))
	require.NoError(t, err)
	require.NotNil(t, m["a"][0].GetTerm1().GetDescription())
	require.Equal(t, "letter a", *m["a"][0].GetTerm1().GetDescription())
}

func TestConfusionSetLoader_VendoredUpstreamEN(t *testing.T) {
	// Load official confusion_sets.txt from testdata/upstream.
	path := findUpstreamConfusionSets(t)
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	m, err := NewConfusionSetLoader(nil).LoadConfusionPairs(f)
	require.NoError(t, err)
	require.Greater(t, len(m), 100, "expected many confusion pairs from upstream")
	// common pair: affect/effect (order alphabetical in file)
	_, ok1 := m["affect"]
	_, ok2 := m["effect"]
	require.True(t, ok1 || ok2, "expected affect/effect confusion pair")
}

func findUpstreamConfusionSets(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	dir := wd
	for {
		cand := filepath.Join(dir, "testdata", "upstream", "en", "resource", "confusion_sets.txt")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		p := filepath.Dir(dir)
		if p == dir {
			t.Fatal("vendored confusion_sets.txt not found")
		}
		dir = p
	}
}
