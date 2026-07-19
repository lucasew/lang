package patterns

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadOfficialFRGrammar_UnifyLoaded(t *testing.T) {
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
	unified := 0
	for _, ar := range ars {
		if ar != nil && ar.TestUnification {
			unified++
		}
	}
	t.Logf("FR grammar rules: %d (with unify: %d)", len(ars), unified)
	require.Greater(t, len(ars), 100)
	require.Greater(t, unified, 0, "FR grammar has many <unify> rules; loader must keep them")
	// Unification equivalence tables from <unification feature="…">
	require.NotNil(t, l.LastUnifierConfig)
	feats := l.LastUnifierConfig.GetEquivalenceFeatures()
	require.Contains(t, feats, "number")
	require.Contains(t, feats, "gender")
}
