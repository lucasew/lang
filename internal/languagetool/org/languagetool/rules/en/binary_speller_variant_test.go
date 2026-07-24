package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishVariantSpellerMeta(t *testing.T) {
	id, f := EnglishVariantSpellerMeta("en-GB")
	require.Equal(t, MorfologikBritishSpellerRuleID, id)
	require.Equal(t, "en_GB.dict", f)
	id, f = EnglishVariantSpellerMeta("en")
	require.Equal(t, MorfologikAmericanSpellerRuleID, id)
	require.Equal(t, "en_US.dict", f)
}

func TestRegisterCore_BinaryBritishSpeller(t *testing.T) {
	p := DiscoverEnglishVariantDictFile("en_GB.dict")
	if p == "" {
		t.Skip("en_GB.dict not in tree")
	}
	lt := languagetool.NewJLanguageTool("en-GB")
	RegisterCoreEnglishLanguageRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), MorfologikBritishSpellerRuleID)
	// misspelling should fire with British rule ID (not invent US-only)
	m := lt.Check("I recieve the book.")
	var found bool
	for _, x := range m {
		if x.RuleID == MorfologikBritishSpellerRuleID {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}

func TestRegisterCore_BinaryAmericanSpeller(t *testing.T) {
	if DiscoverEnglishVariantDictFile("en_US.dict") == "" {
		t.Skip("en_US.dict not in tree")
	}
	lt := languagetool.NewJLanguageTool("en")
	RegisterCoreEnglishLanguageRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), MorfologikAmericanSpellerRuleID)
	m := lt.Check("I recieve the book.")
	var found bool
	for _, x := range m {
		if x.RuleID == MorfologikAmericanSpellerRuleID {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
