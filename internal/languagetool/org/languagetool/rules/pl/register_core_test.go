package pl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCorePolishRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pl")
	RegisterCorePolishRules(lt)
	m := lt.Check("Idę w w domu.")
	found := false
	for _, x := range m {
		if x.RuleID == "PL_W_W" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
