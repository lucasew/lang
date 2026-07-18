package patterns

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadOfficialFRGrammar_UnifySkipped(t *testing.T) {
	cands := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "fr",
			"src", "main", "resources", "org", "languagetool", "rules", "fr", "grammar.xml"),
		filepath.Join("testdata", "upstream", "fr", "rules", "grammar.xml"),
	}
	var p string
	cwd, _ := os.Getwd()
	// walk up from package dir
	dir := cwd
	for i := 0; i < 8 && p == ""; i++ {
		for _, rel := range cands {
			cand := filepath.Join(dir, rel)
			if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
				p = cand
				break
			}
		}
		dir = filepath.Dir(dir)
	}
	if p == "" {
		t.Skip("fr grammar.xml not found")
	}
	data, err := ReadExpandedGrammarFile(p)
	require.NoError(t, err)
	l := NewPatternRuleLoader()
	l.SetRelaxedMode(true)
	ars, err := l.GetRulesFromString(string(data), p, "fr")
	require.NoError(t, err)
	t.Logf("FR grammar surface rules (unify skipped): %d", len(ars))
	// Should still load many non-unify rules
	require.Greater(t, len(ars), 100)
}
