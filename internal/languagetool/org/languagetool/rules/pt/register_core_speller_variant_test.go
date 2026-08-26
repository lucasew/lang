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

func TestRegisterCore_AngolaSpellerID(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-AO")
	RegisterCorePortugueseRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_PT_AO")
	require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_PT_PT")
	require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_PT_BR")
	require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_PT_MZ")
}

func TestRegisterCore_MozambiqueSpellerID(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-MZ")
	RegisterCorePortugueseRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_PT_MZ")
	require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_PT_PT")
	require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_PT_BR")
	require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_PT_AO")
}
