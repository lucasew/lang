package sk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreSlovakRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sk")
	RegisterCoreSlovakRules(lt)
	m := lt.Check("Idem v v dome.")
	found := false
	for _, x := range m {
		if x.RuleID == "SK_V_V" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
