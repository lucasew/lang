package da

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreDanishRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("da")
	RegisterCoreDanishRules(lt)
	m := lt.Check("Han gik i i huset.")
	found := false
	for _, x := range m {
		if x.RuleID == "DA_I_I" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
