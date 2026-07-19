package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("pt")
	RegisterCorePortugueseRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "PT_A_O")
}
