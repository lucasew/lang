package sv

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreSwedishRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sv")
	RegisterCoreSwedishRules(lt)
	m := lt.Check("Han gick i i huset.")
	found := false
	for _, x := range m {
		if x.RuleID == "SV_I_I" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
