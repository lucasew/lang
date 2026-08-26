package ro

// Twin of languagetool-language-modules/ro/src/test/java/org/languagetool/rules/ro/RomanianTest.java
//
// Java testLanguage:
//   runTests(new Romanian()) — LanguageSpecificTest suite (demo text when configured,
//   pattern grammar.xml rules, plus core getRelevantRules).
//
// Full LanguageSpecificTest.runTests may include pattern grammar from grammar.xml
// which is not auto-loaded by NewJLanguageTool("ro") + RegisterCoreRomanianRules
// without activateDefaultPatternRules. This twin asserts Java-visible surface the
// port can prove without invent: metadata, getRelevantRules IDs, and real core-rule
// outcomes (unpaired brackets, whitespace/double punctuation, simple replace,
// speller when dict loads). Honest gap: grammar.xml pattern matches not asserted.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

// Port of RomanianTest.testLanguage — metadata + relevant rules + provable core outcomes.
func TestRomanian_Language(t *testing.T) {
	lang := language.NewRomanian()
	require.Equal(t, "ro", lang.GetShortCode())
	require.Equal(t, "Romanian", lang.GetName())
	require.Equal(t, []string{"RO"}, lang.GetCountries())
	require.Equal(t, languagetool.LookingForNewMaintainer, lang.GetMaintainedState())
	require.Equal(t, "Ionuț Păduraru", lang.GetMaintainers()[0].Name)
	require.Equal(t, "http://www.archeus.ro", lang.GetMaintainers()[0].URL)
	require.Equal(t, "MORFOLOGIK_RULE_RO_RO", lang.GetDefaultSpellingRuleID())

	// getRelevantRules IDs in Java Arrays.asList order
	require.Equal(t, language.RomanianRelevantRuleIDs(), lang.GetRelevantRuleIDs())
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"UNPAIRED_BRACKETS",
		"WORD_REPEAT_RULE",
		"MORFOLOGIK_RULE_RO_RO",
		"ROMANIAN_WORD_REPEAT_BEGINNING_RULE",
		"RO_SIMPLE_REPLACE",
		"RO_COMPOUND",
	}, lang.GetRelevantRuleIDs())

	// Wire Java getRelevantRules into JLanguageTool (no invent extras).
	lt := languagetool.NewJLanguageTool("ro")
	RegisterCoreRomanianRules(lt)
	// Order matches Java getRelevantRules registration order.
	require.Equal(t, language.RomanianRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())

	// Core-rule outcomes (Java-visible without grammar.xml pattern pack):
	// unpaired „ …  → UNPAIRED_BRACKETS
	// double space   → WHITESPACE_RULE
	// "!!"           → DOUBLE_PUNCTUATION
	// "patrusprezece" → RO_SIMPLE_REPLACE
	// deliberate typo → MORFOLOGIK_RULE_RO_RO (when dict loads)
	// "test test"    → WORD_REPEAT_RULE

	// Unpaired brackets (Java GenericUnpairedBracketsRuleTest RO symbols)
	matches := lt.Check("A fost plecat „pentru... puțin timp.")
	ids := matchIDs(matches)
	require.Contains(t, ids, "UNPAIRED_BRACKETS",
		"unpaired „ must match GenericUnpairedBracketsRule (Java UNPAIRED_BRACKETS)")

	// Multiple whitespace
	matches = lt.Check("Aici  sunt  două spații.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "WHITESPACE_RULE",
		"double spaces must match MultipleWhitespaceRule")

	// Double punctuation (Java DoublePunctuationRule: consecutive dots/commas only)
	matches = lt.Check("Text,, greșit.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "DOUBLE_PUNCTUATION",
		",, must match DoublePunctuationRule")

	// Simple replace (Java SimpleReplaceRuleTest)
	matches = lt.Check("Patrusprezece case.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "RO_SIMPLE_REPLACE",
		"'Patrusprezece' must match SimpleReplaceRule")

	// Word repeat
	matches = lt.Check("Acesta este un test test simplu.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "WORD_REPEAT_RULE",
		"'test test' must match WordRepeatRule")

	// Speller — require ≥1 hit when official ro_RO.dict loads
	matches = lt.Check("Aceasta este o greselaa de scriere.")
	ids = matchIDs(matches)
	spellerHits := 0
	for _, id := range ids {
		if id == "MORFOLOGIK_RULE_RO_RO" {
			spellerHits++
		}
	}
	require.GreaterOrEqual(t, spellerHits, 1,
		"deliberate misspelling must hit MorfologikRomanianSpellerRule when ro_RO.dict loads")
}

// Analyze surface (runTests-adjacent) — real RO sentence split.
func TestRomanian_AnalyzeDemoText(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ro")
	sents := lt.Analyze("Acesta este un text de test. Și aceasta este a doua propoziție.")
	require.NotEmpty(t, sents)
	require.GreaterOrEqual(t, len(sents), 2)
}

func matchIDs(matches []languagetool.LocalMatch) []string {
	ids := make([]string, 0, len(matches))
	for _, m := range matches {
		ids = append(ids, m.RuleID)
	}
	return ids
}
