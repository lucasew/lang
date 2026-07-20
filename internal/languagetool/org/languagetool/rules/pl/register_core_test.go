package pl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("pl")
	RegisterCorePolishRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "PL_W_W")
}

// Java Polish.getRelevantRules: WordRepeatRule + PolishWordRepeatRule (no beginning).
func TestRegisterCorePolishRules_BothWordRepeatIDs(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pl")
	RegisterCorePolishRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.Contains(t, ids, "WORD_REPEAT_RULE")
	require.Contains(t, ids, "PL_WORD_REPEAT")
	require.NotContains(t, ids, "PL_WORD_REPEAT_BEGINNING_RULE")
}
