package be

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Official twin pair: кампутар → камп’ютар (SimpleReplaceRuleTest).
func TestRegisterCoreBelarusianRules_Replace(t *testing.T) {
	lt := languagetool.NewJLanguageTool("be")
	RegisterCoreBelarusianRules(lt)
	m := lt.Check("Яго кампутар выключыўся.")
	found := false
	for _, x := range m {
		if x.RuleID == "BE_SIMPLE_REPLACE" {
			found = true
			break
		}
	}
	require.True(t, found, "%+v", m)
}
