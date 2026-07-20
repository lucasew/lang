package is

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Java Icelandic.getRelevantRules: WordRepeatRule id WORD_REPEAT_RULE; no beginning rule.
func TestRegisterCoreIcelandicRules_NoInventBeginning(t *testing.T) {
	lt := languagetool.NewJLanguageTool("is")
	RegisterCoreIcelandicRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.Contains(t, ids, "WORD_REPEAT_RULE")
	require.Contains(t, ids, "HUNSPELL_NO_SUGGEST_RULE")
	require.NotContains(t, ids, "WORD_REPEAT_BEGINNING_RULE")
	require.NotContains(t, ids, "IS_WORD_REPEAT_BEGINNING_RULE")
	require.NotContains(t, ids, "IS_WORD_REPEAT_RULE")
}
