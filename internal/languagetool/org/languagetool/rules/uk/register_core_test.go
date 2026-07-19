package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("uk")
	RegisterCoreUkrainianRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "UK_В_В")
}
