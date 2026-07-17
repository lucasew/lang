package sr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Upstream ekavian replace-grammar.txt: еуро=евро (same pair as twin Java test).
func TestRegisterCoreSerbianRules_ReplaceGrammar(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sr")
	RegisterCoreSerbianRules(lt)
	m := lt.Check("То кошта један еуро.")
	found := false
	for _, x := range m {
		if x.RuleID == "SR_EKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE" {
			found = true
			break
		}
	}
	require.True(t, found, "%+v", m)
}
