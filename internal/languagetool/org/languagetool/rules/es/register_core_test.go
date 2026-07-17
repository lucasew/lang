package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreSpanishRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("es")
	RegisterCoreSpanishRules(lt)
	m := lt.Check("Voy a el mercado.")
	found := false
	for _, x := range m {
		if x.RuleID == "ES_A_EL" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
