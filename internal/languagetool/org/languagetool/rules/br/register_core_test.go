package br

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("br")
	RegisterCoreBretonRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "BR_HA_HA")
}

// Java Breton.getRelevantRules has no word-repeat / compound registration.
func TestRegisterCoreBretonRules_NoInventWordRepeatOrCompound(t *testing.T) {
	lt := languagetool.NewJLanguageTool("br")
	RegisterCoreBretonRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.Contains(t, ids, "BR_TOPO")
	require.Contains(t, ids, "MORFOLOGIK_RULE_BR_FR")
	for _, id := range ids {
		require.NotContains(t, id, "WORD_REPEAT", "not in Java getRelevantRules")
	}
	require.NotContains(t, ids, "BR_COMPOUNDS") // Java does not register BretonCompoundRule in getRelevantRules
	// Bare repeat must not invent a match
	for _, m := range lt.Check("test test") {
		require.NotContains(t, m.RuleID, "WORD_REPEAT")
	}
}
