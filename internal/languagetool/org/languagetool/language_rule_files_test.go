package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRuleFileNames(t *testing.T) {
	exists := map[string]bool{
		"en/grammar.xml":          true,
		"en/style.xml":            true,
		"en/grammar_custom.xml":   false,
		"en/en-US/grammar.xml":    true,
	}
	files := GetRuleFileNames("en", "en-US", "/org/languagetool/rules", func(p string) bool {
		return exists[p]
	})
	require.Contains(t, files, "/org/languagetool/rules/en/grammar.xml")
	require.Contains(t, files, "/org/languagetool/rules/en/style.xml")
	require.Contains(t, files, "/org/languagetool/rules/en/en-US/grammar.xml")
	require.NotContains(t, files, "/org/languagetool/rules/en/grammar_custom.xml")
}
