package en

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadSoftTyposFile(t *testing.T) {
	wd, _ := os.Getwd()
	path := ""
	dir := wd
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "testdata", "spelling", "en-typos.tsv")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			path = cand
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if path == "" {
		t.Skip("en-typos.tsv not found")
	}
	m, err := LoadSoftTyposFile(path)
	require.NoError(t, err)
	require.Contains(t, m["recieve"], "receive")
	require.Contains(t, m["teh"], "the")
	require.Contains(t, m["wierd"], "weird")
	// lower-case key too
	require.Contains(t, m["wierd"], "weird")
}

func TestMergeSpellerSuggestions(t *testing.T) {
	base := map[string][]string{"teh": {"the"}}
	extra := map[string][]string{"teh": {"the", "teh!"}, "foo": {"bar"}}
	m := MergeSpellerSuggestions(base, extra)
	require.Equal(t, []string{"the", "teh!"}, m["teh"])
	require.Equal(t, []string{"bar"}, m["foo"])
}
