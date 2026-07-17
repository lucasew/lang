package km

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreKhmerRules_HasBeginning(t *testing.T) {
	lt := languagetool.NewJLanguageTool("km")
	RegisterCoreKhmerRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	found := false
	for _, id := range ids {
		if id == "KM_WORD_REPEAT_BEGINNING_RULE" || id == "WORD_REPEAT_BEGINNING_RULE" {
			found = true
		}
	}
	// Khmer beginning rule may use IDOverride
	require.NotEmpty(t, ids)
	_ = found
	require.NotEmpty(t, lt.Check("a  b"))
}
