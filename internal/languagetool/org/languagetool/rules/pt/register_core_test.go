package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("pt")
	RegisterCorePortugueseRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "PT_A_O")
}

// uniqueIDs collapses duplicate Java asList / second-instance getId entries.
func uniqueRuleIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// Java PortugalPortuguese.getRelevantRules (default product surface for bare "pt").
func TestRegisterCorePortugueseRules_PortugalPortugueseSurface(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt")
	RegisterCorePortugueseRules(lt)
	want := language.PortugalPortuguese.GetRelevantRuleIDs()
	require.ElementsMatch(t, uniqueRuleIDs(want), uniqueRuleIDs(lt.GetAllRegisteredRuleIDs()))
	// Unique extras after super (PortugalPortuguese.java)
	for _, id := range []string{
		"PT_POSAO_DASH_RULE", "PT_AGREEMENT_REPLACE", "PT_PT_SIMPLE_REPLACE",
		"PT_ARCHAISMS_REPLACE",
	} {
		require.Contains(t, lt.GetAllRegisteredRuleIDs(), id)
	}
	// Invent / wrong IDs / preAO-only extras must not reappear on pt surface
	for _, bad := range []string{
		"WHITESPACE_PUNCTUATION", "PT_UNPAIRED_BRACKETS",
		"PT_COMPOUNDS_PRE_REFORM", "PT_PREAO_DASH_RULE",
		"PT_BR_SIMPLE_REPLACE", "PT_WEASELWORD_REPLACE",
		"MORFOLOGIK_RULE_PT_AO", "MORFOLOGIK_RULE_PT_MZ",
	} {
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), bad)
	}
}

// Java PortugalPortuguese.getRelevantRules for explicit pt-PT.
func TestRegisterCorePortugueseRules_ptPT(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-PT")
	RegisterCorePortugueseRules(lt)
	want := language.PortugalPortuguese.GetRelevantRuleIDs()
	require.ElementsMatch(t, uniqueRuleIDs(want), uniqueRuleIDs(lt.GetAllRegisteredRuleIDs()))
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_PT_PT")
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "PT_PT_SIMPLE_REPLACE")
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "PT_AGREEMENT_REPLACE")
}

// Java BrazilianPortuguese.getRelevantRules for pt-BR.
func TestRegisterCorePortugueseRules_BrazilianPortugueseSurface(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-BR")
	RegisterCorePortugueseRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := language.BrazilianPortuguese.GetRelevantRuleIDs()
	require.ElementsMatch(t, uniqueRuleIDs(want), uniqueRuleIDs(ids))
	require.Contains(t, ids, "MORFOLOGIK_RULE_PT_BR")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_PT")
	// Unique extras after super (BrazilianPortuguese.java)
	for _, id := range []string{
		"PT_POSAO_DASH_RULE", "PT_BR_SIMPLE_REPLACE", "PT_ARCHAISMS_REPLACE",
	} {
		require.Contains(t, ids, id)
	}
	// Portugal-only extras must not leak onto pt-BR.
	require.NotContains(t, ids, "PT_PT_SIMPLE_REPLACE")
	require.NotContains(t, ids, "PT_AGREEMENT_REPLACE")
	// preAO PreReform extras / spellers must not leak onto pt-BR.
	require.NotContains(t, ids, "PT_COMPOUNDS_PRE_REFORM")
	require.NotContains(t, ids, "PT_PREAO_DASH_RULE")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_AO")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_MZ")
}

// Brazilian speller ID still selected for pt-BR (also covered above).
func TestRegisterCorePortugueseRules_BrazilianSpellerID(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-BR")
	RegisterCorePortugueseRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.Contains(t, ids, "MORFOLOGIK_RULE_PT_BR")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_PT")
}

// Java AngolaPortuguese.getRelevantRules for pt-AO.
func TestRegisterCorePortugueseRules_AngolaPortugueseSurface(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-AO")
	RegisterCorePortugueseRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := language.AngolaPortuguese.GetRelevantRuleIDs()
	require.ElementsMatch(t, uniqueRuleIDs(want), uniqueRuleIDs(ids))
	require.Contains(t, ids, "MORFOLOGIK_RULE_PT_AO")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_PT")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_BR")
	// PreReform extras after super (AngolaPortuguese.java)
	for _, id := range []string{
		"PT_COMPOUNDS_PRE_REFORM", "PT_PREAO_DASH_RULE",
	} {
		require.Contains(t, ids, id)
	}
	// PT/BR packs and PostReform dash must not leak onto pt-AO.
	for _, bad := range []string{
		"PT_POSAO_DASH_RULE", "PT_AGREEMENT_REPLACE", "PT_PT_SIMPLE_REPLACE",
		"PT_BR_SIMPLE_REPLACE", "PT_ARCHAISMS_REPLACE", "MORFOLOGIK_RULE_PT_MZ",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Angola speller ID selected for pt-AO.
func TestRegisterCorePortugueseRules_AngolaSpellerID(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-AO")
	RegisterCorePortugueseRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.Contains(t, ids, "MORFOLOGIK_RULE_PT_AO")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_PT")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_BR")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_MZ")
}

// Java MozambiquePortuguese.getRelevantRules for pt-MZ.
func TestRegisterCorePortugueseRules_MozambiquePortugueseSurface(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-MZ")
	RegisterCorePortugueseRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := language.MozambiquePortuguese.GetRelevantRuleIDs()
	require.ElementsMatch(t, uniqueRuleIDs(want), uniqueRuleIDs(ids))
	require.Contains(t, ids, "MORFOLOGIK_RULE_PT_MZ")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_PT")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_BR")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_AO")
	// PreReform extras after super (MozambiquePortuguese.java)
	for _, id := range []string{
		"PT_COMPOUNDS_PRE_REFORM", "PT_PREAO_DASH_RULE",
	} {
		require.Contains(t, ids, id)
	}
	// PT/BR packs and PostReform dash must not leak onto pt-MZ.
	for _, bad := range []string{
		"PT_POSAO_DASH_RULE", "PT_AGREEMENT_REPLACE", "PT_PT_SIMPLE_REPLACE",
		"PT_BR_SIMPLE_REPLACE", "PT_ARCHAISMS_REPLACE",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Mozambique speller ID selected for pt-MZ.
func TestRegisterCorePortugueseRules_MozambiqueSpellerID(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-MZ")
	RegisterCorePortugueseRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.Contains(t, ids, "MORFOLOGIK_RULE_PT_MZ")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_PT")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_BR")
	require.NotContains(t, ids, "MORFOLOGIK_RULE_PT_AO")
}
