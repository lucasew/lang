package pl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestRegisterCore_BinaryPolishSpellerWhenDictPresent(t *testing.T) {
	if morfologik.DiscoverLanguageDict(PolishSpellerDict) == "" {
		t.Skip("pl_PL.dict not in tree")
	}
	lt := languagetool.NewJLanguageTool("pl")
	RegisterCorePolishRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), MorfologikPolishSpellerRuleID)
	m := lt.Check("xyzzyqqq")
	var found bool
	for _, x := range m {
		if x.RuleID == MorfologikPolishSpellerRuleID {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
