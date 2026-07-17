package ro

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRomanianRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ro")
	RegisterCoreRomanianRules(lt)
	m := lt.Check("Am nevoie de de ajutor.")
	found := false
	for _, x := range m {
		if x.RuleID == "RO_DE_DE" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
