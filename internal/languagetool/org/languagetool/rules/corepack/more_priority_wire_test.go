package corepack_test

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"
	"github.com/stretchr/testify/require"
)

func TestRegister_GalicianPriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("gl")
	corepack.Register(lt, "gl")
	require.NotNil(t, lt.PriorityForId)
	require.Equal(t, 30, lt.PriorityForId("DEGREE_MINUTES_SECONDS"))
	require.Equal(t, -1001, lt.PriorityForId("TOO_LONG_SENTENCE_40"))
}

func TestRegister_RussianPriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ru")
	corepack.Register(lt, "ru")
	require.NotNil(t, lt.PriorityForId)
	require.Equal(t, 12, lt.PriorityForId("RU_DASH_RULE"))
	require.Equal(t, 1, lt.PriorityForId("MORFOLOGIC_RULE_RU_RU"))
}

func TestRegister_BelarusianPriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("be")
	corepack.Register(lt, "be")
	require.NotNil(t, lt.PriorityForId)
	require.Equal(t, 9, lt.PriorityForId("BELARUSIAN_SPECIFIC_CASE"))
}

func TestRegister_IrishPriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ga")
	corepack.Register(lt, "ga")
	require.NotNil(t, lt.PriorityForId)
	require.Equal(t, -15, lt.PriorityForId("TOO_LONG_PARAGRAPH"))
}

func TestRegister_PolishPriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pl")
	corepack.Register(lt, "pl")
	require.NotNil(t, lt.PriorityForId)
	require.Equal(t, -1, lt.PriorityForId("ZDANIA_ZLOZONE"))
}

func TestRegister_SimpleGermanPriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de-DE-x-simple-language")
	corepack.Register(lt, "de")
	require.NotNil(t, lt.PriorityForId)
	require.Equal(t, 10, lt.PriorityForId("TOO_LONG_SENTENCE"))
	require.Equal(t, -1, lt.PriorityForId("LANGES_WORT"))
	// still German super for map ids
	require.Equal(t, 10, lt.PriorityForId("OLD_SPELLING_RULE"))
}

func TestRegister_EnglishDefaultStylePriority(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	corepack.Register(lt, "en")
	require.Equal(t, -50, lt.DefaultRulePriorityForStyle)
}

func TestRegister_CatalanDefaultStylePriority(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ca")
	corepack.Register(lt, "ca")
	require.Equal(t, -50, lt.DefaultRulePriorityForStyle)
}

func TestRegister_RussianIgnoredCharacters(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ru")
	corepack.Register(lt, "ru")
	require.NotNil(t, lt.IgnoredCharacters)
}

func TestRegister_UkrainianIgnoredCharacters(t *testing.T) {
	lt := languagetool.NewJLanguageTool("uk")
	corepack.Register(lt, "uk")
	require.NotNil(t, lt.IgnoredCharacters)
}

func TestRegister_BelarusianIgnoredCharacters(t *testing.T) {
	lt := languagetool.NewJLanguageTool("be")
	corepack.Register(lt, "be")
	require.NotNil(t, lt.IgnoredCharacters)
}

func TestRegister_HunspellSpellersWired(t *testing.T) {
	// Java default/relevant spellers: HunspellRule or HunspellNoSuggestionRule ids.
	for _, code := range []string{"da", "sv", "gl", "eo"} {
		lt := languagetool.NewJLanguageTool(code)
		corepack.Register(lt, code)
		ids := lt.GetAllRegisteredRuleIDs()
		require.Contains(t, ids, "HUNSPELL_RULE", code)
		require.NotContains(t, ids, "MORFOLOGIK_RULE_DA_DK", code)
	}
	lt := languagetool.NewJLanguageTool("is")
	corepack.Register(lt, "is")
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "HUNSPELL_NO_SUGGEST_RULE")
}
