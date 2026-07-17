package el

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreGreekRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("el")
	RegisterCoreGreekRules(lt)
	m := lt.Check("Ήρθε και και έφυγε.")
	found := false
	for _, x := range m {
		if x.RuleID == "EL_ΚΑΙ_ΚΑΙ" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
