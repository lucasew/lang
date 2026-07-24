package morfologik

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDiscoverLanguageDict_Polish(t *testing.T) {
	p := DiscoverLanguageDict("/pl/hunspell/pl_PL.dict")
	if p == "" {
		t.Skip("pl_PL.dict not in tree")
	}
	require.Contains(t, p, "pl_PL.dict")
}

func TestTryRegisterBinarySpeller_Polish(t *testing.T) {
	if DiscoverLanguageDict("/pl/hunspell/pl_PL.dict") == "" {
		t.Skip("pl_PL.dict not in tree")
	}
	lt := languagetool.NewJLanguageTool("pl")
	require.True(t, TryRegisterBinarySpeller(lt, "MORFOLOGIK_RULE_PL_PL", "/pl/hunspell/pl_PL.dict"))
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_PL_PL")
	// obvious misspelling — Polish word form not needed for flag path
	m := lt.Check("xyzzyqqq")
	var found bool
	for _, x := range m {
		if x.RuleID == "MORFOLOGIK_RULE_PL_PL" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
