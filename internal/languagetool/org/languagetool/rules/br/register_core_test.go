package br

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreBretonRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("br")
	RegisterCoreBretonRules(lt)
	m := lt.Check("Bras ha ha bihan.")
	found := false
	for _, x := range m {
		if x.RuleID == "BR_HA_HA" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
