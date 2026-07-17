package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCorePortugueseRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt")
	RegisterCorePortugueseRules(lt)
	m := lt.Check("Vou a o mercado.")
	found := false
	for _, x := range m {
		if x.RuleID == "PT_A_O" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
