package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("pt")
	RegisterCorePortugueseRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "PT_A_O")
}

// Java Portuguese.getRelevantRules exact ID set (pt-PT speller).
func TestRegisterCorePortugueseRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt")
	RegisterCorePortugueseRules(lt)
	require.ElementsMatch(t, language.PortugueseRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())
	for _, bad := range []string{
		"WHITESPACE_PUNCTUATION", "PT_UNPAIRED_BRACKETS",
		"PT_AGREEMENT_REPLACE", "PT_COMPOUNDS_PRE_REFORM", "PT_PREAO_DASH_RULE",
		"PT_POSAO_DASH_RULE", "PT_PT_SIMPLE_REPLACE", "PT_BR_SIMPLE_REPLACE",
		"PT_ARCHAISMS_REPLACE", "PT_WEASELWORD_REPLACE",
	} {
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), bad)
	}
}

// Brazilian variant uses MORFOLOGIK_RULE_PT_BR.
func TestRegisterCorePortugueseRules_BrazilianSpellerID(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-BR")
	RegisterCorePortugueseRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.Contains(t, ids, "MORFOLOGIK_RULE_PT_BR")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_PT")
}
