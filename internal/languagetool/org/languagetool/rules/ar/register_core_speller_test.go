package ar

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Faithful: RegisterCore registers Java ArabicHunspellSpellerRule getId.
func TestRegisterCore_ArabicHunspellSpellerID(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ar")
	RegisterCoreArabicRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "HUNSPELL_RULE_AR")
}
