package km

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"
	"github.com/stretchr/testify/require"
)

// Faithful: RegisterCore registers Java KhmerHunspellRule (extends HunspellRule → HUNSPELL_RULE).
func TestRegisterCore_KhmerHunspellSpellerID(t *testing.T) {
	lt := languagetool.NewJLanguageTool("km")
	RegisterCoreKhmerRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), hunspell.HunspellRuleID)
}
