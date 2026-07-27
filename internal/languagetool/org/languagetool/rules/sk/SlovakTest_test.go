package sk

// Twin of languagetool-language-modules/sk/src/test/java/org/languagetool/rules/sk/SlovakTest.java
//
// Java testLanguage:
//   // demo string commented out (outdated)
//   runTests(new Slovak()) — LanguageSpecificTest suite (demo text when configured,
//   pattern grammar.xml rules, plus core getRelevantRules).
//
// Full LanguageSpecificTest.runTests may include pattern grammar from grammar.xml
// / grammar-typography.xml which is not auto-loaded by NewJLanguageTool("sk") +
// RegisterCoreSlovakRules without activateDefaultPatternRules. This twin asserts
// Java-visible surface the port can prove without invent: metadata, getRelevantRules
// IDs, and real core-rule outcomes (unpaired brackets, whitespace/double punctuation,
// word repeat, speller when dict loads). Honest gap: grammar.xml / grammar-typography.xml
// pattern matches not asserted.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of SlovakTest.testLanguage — metadata + relevant rules + provable core outcomes.
func TestSlovak_Language(t *testing.T) {
	lang := language.NewSlovak()
	require.Equal(t, "sk", lang.GetShortCode())
	require.Equal(t, "Slovak", lang.GetName())
	require.Equal(t, []string{"SK"}, lang.GetCountries())
	require.Equal(t, languagetool.LookingForNewMaintainer, lang.GetMaintainedState())
	require.Equal(t, "Zdenko Podobný", lang.GetMaintainers()[0].Name)
	require.Equal(t, "http://sk-spell.sk.cx", lang.GetMaintainers()[0].URL)
	require.Equal(t, "MORFOLOGIK_RULE_SK_SK", lang.GetDefaultSpellingRuleID())

	// getRelevantRules IDs in Java Arrays.asList order
	require.Equal(t, language.SlovakRelevantRuleIDs(), lang.GetRelevantRuleIDs())
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"UPPERCASE_SENTENCE_START",
		"WORD_REPEAT_RULE",
		"WHITESPACE_RULE",
		"SK_COMPOUNDS",
		"MORFOLOGIK_RULE_SK_SK",
	}, lang.GetRelevantRuleIDs())

	// Wire Java getRelevantRules into JLanguageTool (no invent extras).
	lt := languagetool.NewJLanguageTool("sk")
	RegisterCoreSlovakRules(lt)
	// Order matches Java getRelevantRules registration order.
	require.Equal(t, language.SlovakRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())

	// Core-rule outcomes (Java-visible without grammar.xml pattern pack):
	// unpaired „ …  → UNPAIRED_BRACKETS
	// double space   → WHITESPACE_RULE
	// ",,"           → DOUBLE_PUNCTUATION
	// "test test"    → WORD_REPEAT_RULE
	// deliberate typo → MORFOLOGIK_RULE_SK_SK (when dict loads)

	// Unpaired brackets (Java Slovak GenericUnpairedBracketsRule symbols)
	matches := lt.Check("Toto je text „bez uzavretia.")
	ids := matchIDs(matches)
	require.Contains(t, ids, "UNPAIRED_BRACKETS",
		"unpaired „ must match GenericUnpairedBracketsRule (Java UNPAIRED_BRACKETS)")

	// Multiple whitespace
	matches = lt.Check("Tu  sú  dve medzery.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "WHITESPACE_RULE",
		"double spaces must match MultipleWhitespaceRule")

	// Double punctuation (Java DoublePunctuationRule: consecutive dots/commas only)
	matches = lt.Check("Text,, chybný.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "DOUBLE_PUNCTUATION",
		",, must match DoublePunctuationRule")

	// Word repeat
	matches = lt.Check("Toto je test test jednoduchý.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "WORD_REPEAT_RULE",
		"'test test' must match WordRepeatRule")

	// Speller — require ≥1 hit when official sk_SK.dict loads
	if morfologik.DiscoverLanguageDict(SlovakSpellerDict) == "" {
		t.Log("sk_SK.dict not discoverable — speller outcome skipped (honest gap for missing resource)")
	} else {
		matches = lt.Check("Toto je slovenskyy preklep.")
		ids = matchIDs(matches)
		spellerHits := 0
		for _, id := range ids {
			if id == "MORFOLOGIK_RULE_SK_SK" {
				spellerHits++
			}
		}
		require.GreaterOrEqual(t, spellerHits, 1,
			"deliberate misspelling must hit MorfologikSlovakSpellerRule when sk_SK.dict loads")
	}
}

// Analyze surface (runTests-adjacent) — real SK sentence split.
// Java demo string is commented out; use a short multi-sentence sample instead.
func TestSlovak_AnalyzeDemoText(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sk")
	sents := lt.Analyze("Toto je testovací text. A toto je druhá veta.")
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
