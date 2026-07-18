package patterns

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCombineTogetherSuggestionExpand(t *testing.T) {
	wd, _ := os.Getwd()
	root := wd
	for {
		if st, err := os.Stat(filepath.Join(root, "testdata", "grammar")); err == nil && st.IsDir() {
			break
		}
		root = filepath.Dir(root)
	}
	path := filepath.Join(root, "testdata/grammar/en-upstream-soft.xml")
	lt := languagetool.NewJLanguageTool("en")
	n, err := RegisterGrammarFile(lt, path, "en")
	require.NoError(t, err)
	require.Greater(t, n, 0)
	ms := lt.Check("Two things are combined together in this application.")
	found := false
	for _, m := range ms {
		if m.RuleID != "COMBINE_TOGETHER" {
			continue
		}
		found = true
		t.Logf("msg=%q sug=%v from=%d to=%d", m.Message, m.Suggestions, m.FromPos, m.ToPos)
		require.NotEmpty(t, m.Suggestions)
		require.Equal(t, "combined", m.Suggestions[0])
	}
	require.True(t, found)
}
