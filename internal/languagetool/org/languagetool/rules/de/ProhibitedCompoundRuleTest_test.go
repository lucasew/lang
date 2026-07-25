package de

// Twin of languagetool-language-modules/de/.../ProhibitedCompoundRuleTest.java
//
// Java setup:
//   Map FakeLanguageModel counts; JLanguageTool(de-DE);
//   new ProhibitedCompoundRule(getEnglishMessages(), new FakeLanguageModel(map), null, language)
//
// Active tests: testRule, testRemoveHyphensAndAdaptCase
// Skipped (Java @Ignore): testListOfWords, testMoreThanOneCandidate
//
// Go: de-DE JLanguageTool + WireGermanTagWord (createDefaultTagger) + official de_DE.dict
// for isMisspelled — same FakeLanguageModel map as Java. No AnalyzePlain invent for match matrix.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// javaFakeLMMap ports ProhibitedCompoundRuleTest static map.
func javaFakeLMMap() map[string]int64 {
	return map[string]int64{
		"Mietauto":          100,
		"Leerzeile":         100,
		"Urberliner":        100,
		"Ureinwohner":       100,
		"Wohnungsleerstand": 50,
		"Xliseihflehrstand": 50,
		"Eisensande":        100,
		"Eisenstange":       101,
	}
}

// newProhibitedCompoundFixture mirrors Java ProhibitedCompoundRuleTest fields.
func newProhibitedCompoundFixture(t *testing.T) (*ProhibitedCompoundRule, *languagetool.JLanguageTool) {
	t.Helper()
	// Java: new JLanguageTool(Languages.getLanguageForShortCode("de-DE"))
	lt := languagetool.NewJLanguageTool("de-DE")
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for ProhibitedCompoundRule twin")
	// Java: GermanyGerman.getDefaultSpellingRule().isMisspelled via FilterDict
	if !FilterDictAvailable() {
		p := DiscoverGermanHunspellDict("DE")
		require.NotEmpty(t, p, "de_DE.dict required for isMisspelled twin")
		require.True(t, WireGermanFilterSpeller(p), "wire de_DE.dict")
	}
	// Java: new ProhibitedCompoundRule(getEnglishMessages(), new FakeLanguageModel(map), null, language)
	rule := NewProhibitedCompoundRuleWithFrequency(nil, javaFakeLMMap())
	return rule, lt
}

// TestProhibitedCompoundRule_Rule ports ProhibitedCompoundRuleTest.testRule full matrix.
func TestProhibitedCompoundRule_Rule(t *testing.T) {
	rule, lt := newProhibitedCompoundFixture(t)

	assertM := func(input, expectedMarked, expectedSugg string) {
		t.Helper()
		sent := lt.GetAnalyzedSentence(input)
		require.NotNil(t, sent, "GetAnalyzedSentence(%q)", input)
		ms := rule.Match(sent)
		require.Equal(t, 1, len(ms), "Got matches for %q: %v", input, ms)
		marked := substringByUTF16(input, ms[0].GetFromPos(), ms[0].GetToPos())
		require.Equal(t, expectedMarked, marked, "marked text for %q from=%d to=%d", input, ms[0].GetFromPos(), ms[0].GetToPos())
		require.NotEmpty(t, ms[0].GetSuggestedReplacements())
		// Java: matches[0].getSuggestedReplacements().toString() is "[sugg]"
		require.Equal(t, expectedSugg, ms[0].GetSuggestedReplacements()[0], "suggestion for %q", input)
		require.Equal(t, []string{expectedSugg}, ms[0].GetSuggestedReplacements())
	}
	assert0 := func(input string) {
		t.Helper()
		sent := lt.GetAnalyzedSentence(input)
		require.NotNil(t, sent, "GetAnalyzedSentence(%q)", input)
		ms := rule.Match(sent)
		require.Equal(t, 0, len(ms), "Got matches for %q: %v", input, ms)
	}

	// Order and strings match Java ProhibitedCompoundRuleTest.testRule.
	assertM("Er ist Uhrberliner.", "Uhrberliner", "Urberliner")
	assertM("Er ist Uhr-Berliner.", "Uhr-Berliner", "Urberliner")
	assertM("Das ist ein Mitauto.", "Mitauto", "Mietauto")
	assertM("Das ist ein Mit-Auto.", "Mit-Auto", "Mietauto")
	assert0("Das ist Herr Mitauto.")
	assertM("Hier leben die Uhreinwohner.", "Uhreinwohner", "Ureinwohner")
	assertM("Hier leben die Uhr-Einwohner.", "Uhr-Einwohner", "Ureinwohner")
	assert0("Eine Leerzeile einfügen.")
	assert0("Eine Leer-Zeile einfügen.")
	assertM("Eine Lehrzeile einfügen.", "Lehrzeile", "Leerzeile")
	assertM("Eine Lehr-Zeile einfügen.", "Lehr-Zeile", "Leerzeile")

	assert0("Viel Wohnungsleerstand.")
	assert0("Viel Wohnungs-Leerstand.")
	assertM("Viel Wohnungslehrstand.", "Wohnungslehrstand", "Wohnungsleerstand")
	assertM("Viel Wohnungs-Lehrstand.", "Wohnungs-Lehrstand", "Wohnungsleerstand")
	assert0("Viel Xliseihfleerstand.")
	assert0("Viel Xliseihflehrstand.") // no correct spelling, so not suggested
	assert0("Ein kosmografischer Test")
	assert0("Ein Elektrokardiograph")
	assert0("Die Elektrokardiographen")

	assertM("Den Lehrzeile-Test einfügen.", "Lehrzeile", "Leerzeile")
	assertM("Die Test-Lehrzeile einfügen.", "Lehrzeile", "Leerzeile")
	assertM("Die Versuchs-Test-Lehrzeile einfügen.", "Lehrzeile", "Leerzeile")
	assertM("Den Versuchs-Lehrzeile-Test einfügen.", "Lehrzeile", "Leerzeile")

	// Java SpecificIdRule: match.rule.getId() is toId(DE_PROHIBITED_COMPOUNDS_part1_part2)
	ms := rule.Match(lt.GetAnalyzedSentence("Eine Lehrzeile einfügen."))
	require.Len(t, ms, 1)
	idRule, ok := ms[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "match.rule must be SpecificIdRule")
	require.Contains(t, idRule.GetID(), "DE_PROHIBITED_COMPOUNDS_")
	require.NotEqual(t, "DE_PROHIBITED_COMPOUNDS", idRule.GetID())
	require.Contains(t, idRule.GetDescription(), "Teilwort")
	require.Empty(t, ms[0].GetShortMessage())
}

// TestProhibitedCompoundRule_RemoveHyphensAndAdaptCase ports testRemoveHyphensAndAdaptCase.
func TestProhibitedCompoundRule_RemoveHyphensAndAdaptCase(t *testing.T) {
	rule, _ := newProhibitedCompoundFixture(t)
	// Java assertNull → ""
	require.Equal(t, "", rule.RemoveHyphensAndAdaptCase("Marathonläuse"))
	require.Equal(t, "Marathonläuse", rule.RemoveHyphensAndAdaptCase("Marathon-Läuse"))
	require.Equal(t, "Marathonläusetest", rule.RemoveHyphensAndAdaptCase("Marathon-Läuse-Test"))
	require.Equal(t, "Marathonläusetest", rule.RemoveHyphensAndAdaptCase("Marathon-läuse-test"))
	require.Equal(t, "vieleläusetest", rule.RemoveHyphensAndAdaptCase("viele-Läuse-Test"))
	require.Equal(t, "", rule.RemoveHyphensAndAdaptCase("S-Bahn"))
}

// TestProhibitedCompoundRule_NoFrequencyFailClosed: without LM, Match fails closed
// (Java always constructs with LanguageModel; no Prefer invent).
func TestProhibitedCompoundRule_NoFrequencyFailClosed(t *testing.T) {
	rule := NewProhibitedCompoundRule(nil)
	lt := languagetool.NewJLanguageTool("de-DE")
	_ = WireGermanTagWord(lt)
	require.Empty(t, rule.Match(lt.GetAnalyzedSentence("Da steht eine Lehrzeile zu viel.")))
	require.Empty(t, rule.Match(lt.GetAnalyzedSentence("Das ist ein Mitauto.")))
}

// TestProhibitedCompoundResources_Load: official compound_exceptions + confusion_sets.
func TestProhibitedCompoundResources_Load(t *testing.T) {
	ex := ProhibitedCompoundExceptions()
	require.NotEmpty(t, ex)
	require.Contains(t, ex, "Kostenzeile")
	require.Contains(t, ex, "Kornwinkel")

	pairs := AllProhibitedPairs()
	// base pairs * ~2 case variants + confusion
	require.Greater(t, len(pairs), len(lowercaseProhibitedPairs))
	// confusion_sets: Java only imports pairs where both terms start uppercase
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

	// Aho-Corasick twin: stems + pairMap populated
	stems := prohibitedStems()
	require.NotEmpty(t, stems)
	require.Contains(t, stems, "lehr")
	require.Contains(t, stems, "leer")
	require.NotEmpty(t, prohibitedPairLookup("lehr"))
	require.NotEmpty(t, prohibitedPairLookup("mit"))

	rule := NewProhibitedCompoundRule(nil)
	require.True(t, rule.isBlacklistedWord("Kostenzeile"))
}

// TestProhibitedCompoundRule_Metadata: id, description, example pair (Java ctor).
func TestProhibitedCompoundRule_Metadata(t *testing.T) {
	rule := NewProhibitedCompoundRule(nil)
	require.Equal(t, "DE_PROHIBITED_COMPOUNDS", rule.GetID())
	require.Contains(t, rule.GetDescription(), "Lehrzeile")
	require.NotNil(t, rule.Category)
	require.Len(t, rule.GetIncorrectExamples(), 1)
	require.Len(t, rule.GetCorrectExamples(), 1)
}
