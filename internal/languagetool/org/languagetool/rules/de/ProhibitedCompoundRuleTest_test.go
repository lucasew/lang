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

// Twin of FakeLanguageModel path in ProhibitedCompoundRuleTest.testRule.
func TestProhibitedCompoundRule_WithFrequency(t *testing.T) {
	freq := map[string]int64{
		"Mietauto":          100,
		"Leerzeile":         100,
		"Urberliner":        100,
		"Ureinwohner":       100,
		"Wohnungsleerstand": 50,
		"Xliseihflehrstand": 50,
		"Eisensande":        100,
		"Eisenstange":       101,
	}
	rule := NewProhibitedCompoundRuleWithFrequency(nil, freq)
	// Java isMisspelled(variant): accept only known-correct forms (LM keys that are good)
	rule.IsMisspelled = func(w string) bool {
		ok := map[string]bool{
			"Mietauto": true, "Leerzeile": true, "Urberliner": true, "Ureinwohner": true,
			"Wohnungsleerstand": true, "Eisenstange": true,
		}
		return !ok[w]
	}
	assertM := func(text, sugg string) {
		t.Helper()
		ms := rule.Match(languagetool.AnalyzePlain(text))
		require.Equal(t, 1, len(ms), "text %q", text)
		require.Equal(t, sugg, ms[0].GetSuggestedReplacements()[0], "text %q", text)
	}
	assert0 := func(text string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(text))), "text %q", text)
	}

	assertM("Er ist Uhrberliner.", "Urberliner")
	assertM("Er ist Uhr-Berliner.", "Urberliner")
	assertM("Das ist ein Mitauto.", "Mietauto")
	assertM("Das ist ein Mit-Auto.", "Mietauto")
	assert0("Das ist Herr Mitauto.")
	assertM("Hier leben die Uhreinwohner.", "Ureinwohner")
	assertM("Hier leben die Uhr-Einwohner.", "Ureinwohner")
	assert0("Eine Leerzeile einfügen.")
	assert0("Eine Leer-Zeile einfügen.")
	assertM("Eine Lehrzeile einfügen.", "Leerzeile")
	assertM("Eine Lehr-Zeile einfügen.", "Leerzeile")
	assert0("Viel Wohnungsleerstand.")
	assert0("Viel Wohnungs-Leerstand.")
	assertM("Viel Wohnungslehrstand.", "Wohnungsleerstand")
	assertM("Viel Wohnungs-Lehrstand.", "Wohnungsleerstand")
	assert0("Viel Xliseihfleerstand.")
	assert0("Viel Xliseihflehrstand.") // no correct spelling → not suggested
	assert0("Ein kosmografischer Test")
	assert0("Ein Elektrokardiograph")
	assert0("Die Elektrokardiographen")
	assertM("Den Lehrzeile-Test einfügen.", "Leerzeile")
	assertM("Die Test-Lehrzeile einfügen.", "Leerzeile")
	assertM("Die Versuchs-Test-Lehrzeile einfügen.", "Leerzeile")
	assertM("Den Versuchs-Lehrzeile-Test einfügen.", "Leerzeile")

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
