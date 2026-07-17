package ga

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreIrishRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ga")
	RegisterCoreIrishRules(lt)
	m := lt.Check("Bhí sé agus agus mé.")
	found := false
	for _, x := range m {
		if x.RuleID == "GA_AGUS_AGUS" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
