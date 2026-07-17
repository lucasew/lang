package sl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreSlovenianRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sl")
	RegisterCoreSlovenianRules(lt)
	m := lt.Check("Grem v v hišo.")
	found := false
	for _, x := range m {
		if x.RuleID == "SL_V_V" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
