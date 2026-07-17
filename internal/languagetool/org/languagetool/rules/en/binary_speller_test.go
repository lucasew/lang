package en

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func findEnUSDict(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	dir := wd
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "hunspell", "en_US.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("en_US.dict not found")
	return ""
}

func TestRegisterBinaryEnglishSpeller(t *testing.T) {
	p := findEnUSDict(t)
	lt := languagetool.NewJLanguageTool("en")
	require.True(t, RegisterBinaryEnglishSpeller(lt, p, DemoEnglishKnownWords()))
	m := lt.Check("I recieve teh book.")
	require.NotEmpty(t, m)
	var hasSpell bool
	for _, x := range m {
		if x.RuleID == "MORFOLOGIK_RULE_EN_US" {
			hasSpell = true
		}
	}
	require.True(t, hasSpell, "%+v", m)
	// known word not flagged as spelling
	m = lt.Check("I receive the book.")
	for _, x := range m {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", x.RuleID)
	}
}
