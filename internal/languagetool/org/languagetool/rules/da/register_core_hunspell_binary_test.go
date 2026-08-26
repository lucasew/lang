package da

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"
	"github.com/stretchr/testify/require"
)

func TestRegisterCore_DanishHunspellBinary(t *testing.T) {
	if hunspell.DiscoverHunspellDic(DanishHunspellClasspath) == "" {
		t.Skip("da_DK.dic not in tree")
	}
	lt := languagetool.NewJLanguageTool("da")
	RegisterCoreDanishRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), hunspell.HunspellRuleID)
	m := lt.Check("xyzzyqqqnotaword")
	var found bool
	for _, x := range m {
		if x.RuleID == hunspell.HunspellRuleID {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
