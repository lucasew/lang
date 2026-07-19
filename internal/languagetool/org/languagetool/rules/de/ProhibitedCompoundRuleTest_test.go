package de

// Twin of ProhibitedCompoundRuleTest — Java requires LanguageModel (no Prefer invent).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestProhibitedCompoundRule_Rule(t *testing.T) {
	// Without Frequency, Match fails closed (Java always has LM)
	rule := NewProhibitedCompoundRule(nil)
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Da steht eine Lehrzeile zu viel.")))
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Das ist ein Mitauto.")))
}

func TestProhibitedCompoundRule_RemoveHyphensAndAdaptCase(t *testing.T) {
	require.Equal(t, "", removeHyphensAndAdaptCase("Marathonläuse"))
	require.Equal(t, "Marathonläuse", removeHyphensAndAdaptCase("Marathon-Läuse"))
	require.Equal(t, "Marathonläusetest", removeHyphensAndAdaptCase("Marathon-Läuse-Test"))
	require.Equal(t, "", removeHyphensAndAdaptCase("S-Bahn"))
}

// Twin of FakeLanguageModel path in ProhibitedCompoundRuleTest.
func TestProhibitedCompoundRule_WithFrequency(t *testing.T) {
	freq := map[string]int64{
		"Mietauto":          100,
		"Leerzeile":         100,
		"Urberliner":        100,
		"Ureinwohner":       100,
		"Wohnungsleerstand": 50,
		"Xliseihflehrstand": 50,
	}
	rule := NewProhibitedCompoundRuleWithFrequency(nil, freq)
	// misspelled variants must not be suggested — force misspelled gate off for known goods
	rule.IsMisspelled = func(w string) bool {
		// only accept LM-known forms as correctly spelled
		_, ok := freq[w]
		return !ok
	}
	matchN := func(s string) int { return len(rule.Match(languagetool.AnalyzePlain(s))) }
	require.Equal(t, 1, matchN("Er ist Uhrberliner."))
	require.Equal(t, 1, matchN("Das ist ein Mitauto."))
	require.Equal(t, 0, matchN("Das ist Herr Mitauto."))
	require.Equal(t, 0, matchN("Eine Leerzeile einfügen."))
	require.Equal(t, 1, matchN("Eine Lehrzeile einfügen."))
	require.Equal(t, 0, matchN("Viel Xliseihflehrstand.")) // variant misspelled / not in LM

	// Java SpecificIdRule: match.rule.getId() is toId(DE_PROHIBITED_COMPOUNDS_part1_part2)
	ms := rule.Match(languagetool.AnalyzePlain("Eine Lehrzeile einfügen."))
	require.Len(t, ms, 1)
	idRule, ok := ms[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "match.rule must be SpecificIdRule")
	require.Contains(t, idRule.GetID(), "DE_PROHIBITED_COMPOUNDS_")
	require.NotEqual(t, "DE_PROHIBITED_COMPOUNDS", idRule.GetID())
	require.Contains(t, idRule.GetDescription(), "Teilwort")
	require.Empty(t, ms[0].GetShortMessage())
}

func TestProhibitedCompoundResources_Load(t *testing.T) {
	ex := ProhibitedCompoundExceptions()
	require.NotEmpty(t, ex)
	require.Contains(t, ex, "Kostenzeile")
	require.Contains(t, ex, "Kornwinkel")

	pairs := AllProhibitedPairs()
	// base pairs * ~2 case variants + confusion
	require.Greater(t, len(pairs), len(lowercaseProhibitedPairs))
	// confusion_sets: Java only imports pairs where both terms start uppercase (e.g. Stich/Strich)
	found := false
	for _, p := range pairs {
		if (p.part1 == "Stich" || p.part1 == "stich") && (p.part2 == "Strich" || p.part2 == "strich") {
			found = true
			break
		}
		if (p.part2 == "Stich" || p.part2 == "stich") && (p.part1 == "Strich" || p.part1 == "strich") {
			found = true
			break
		}
	}
	require.True(t, found, "confusion_sets uppercase pairs should be loaded")

	rule := NewProhibitedCompoundRule(nil)
	require.True(t, rule.isBlacklistedWord("Kostenzeile"))
}
