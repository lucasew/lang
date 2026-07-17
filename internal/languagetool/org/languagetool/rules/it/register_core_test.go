package it

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreItalianRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("it")
	RegisterCoreItalianRules(lt)
	m := lt.Check("Vado a il mercato.")
	found := false
	for _, x := range m {
		if x.RuleID == "IT_A_IL" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
