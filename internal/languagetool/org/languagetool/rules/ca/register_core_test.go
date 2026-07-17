package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreCatalanRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ca")
	RegisterCoreCatalanRules(lt)
	m := lt.Check("Vaig a el mercat.")
	found := false
	for _, x := range m {
		if x.RuleID == "CA_A_EL" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
