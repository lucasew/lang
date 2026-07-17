package fa

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCorePersianRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("fa")
	RegisterCorePersianRules(lt)
	m := lt.Check("رفتم در در خانه.")
	found := false
	for _, x := range m {
		if x.RuleID == "FA_در_در" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
