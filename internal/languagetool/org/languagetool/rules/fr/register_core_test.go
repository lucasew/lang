package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreFrenchRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("fr")
	RegisterCoreFrenchRules(lt)
	m := lt.Check("Il est venu malgré que ce soit difficile.")
	found := false
	for _, x := range m {
		if x.RuleID == "FR_MALGRE_QUE" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
