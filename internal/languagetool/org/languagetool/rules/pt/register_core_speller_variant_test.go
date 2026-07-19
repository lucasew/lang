package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCore_BrazilianSpellerID(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-BR")
	RegisterCorePortugueseRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_PT_BR")
	require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_PT_PT")
}
