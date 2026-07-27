package fa

// Twin of languagetool-language-modules/fa/src/test/java/org/languagetool/rules/fa/PersianTest.java
//
// Java testLanguage:
//   demo text commented out (many false alarms on pattern rules)
//   runTests(lang, null, "ā") — LanguageSpecificTest suite
//
// Pattern IDs (Bad_ZWNJ, PluralFix, Complex_Present_Verbs, …) come from fa/grammar.xml
// and are not auto-loaded by NewJLanguageTool("fa") + RegisterCorePersianRules without
// activateDefaultPatternRules. This twin asserts Java-visible surface the port can
// prove without invent: metadata, getRelevantRules IDs, and real core-rule outcomes
// (whitespace/double punctuation, word repeat, simple replace). Honest gap: grammar.xml
// pattern matches not asserted (same class as AST/BE/TL/SL); Java demo text itself is
// commented out upstream.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

// Port of PersianTest.testLanguage — metadata + relevant rules + provable core outcomes.
func TestPersian_Language(t *testing.T) {
	lang := language.NewPersian()
	require.Equal(t, "fa", lang.GetShortCode())
	require.Equal(t, "Persian", lang.GetName())
	require.Equal(t, []string{"IR", "AF"}, lang.GetCountries())
	// multi-country → bare short code (Java buildShortCodeWithCountryAndVariant)
	require.Equal(t, "fa", lang.GetShortCodeWithCountryAndVariant())
	require.Equal(t, languagetool.LookingForNewMaintainer, lang.GetMaintainedState())
	ms := lang.GetMaintainers()
	require.Len(t, ms, 3)
	require.Equal(t, "Reza1615", ms[0].Name)
	require.Equal(t, "Alireza Eskandarpour Shoferi", ms[1].Name)
	require.Equal(t, "Ebrahim Byagowi", ms[2].Name)
	// No invent: no createDefaultSpellingRule
	require.Empty(t, lang.GetDefaultSpellingRuleID())

	// getRelevantRules IDs in Java Arrays.asList order (11 rules)
	require.Equal(t, language.PersianRelevantRuleIDs(), lang.GetRelevantRuleIDs())
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"WHITESPACE_RULE",
		"TOO_LONG_SENTENCE",
		"PERSIAN_COMMA_PARENTHESIS_WHITESPACE",
		"PERSIAN_DOUBLE_PUNCTUATION",
		"PERSIAN_WORD_REPEAT_BEGINNING_RULE",
		"PERSIAN_WORD_REPEAT_RULE",
		"FA_SIMPLE_REPLACE",
		"FA_SPACE_BEFORE_CONJUNCTION",
		"FA_WORD_COHERENCY",
	}, lang.GetRelevantRuleIDs())

	// Wire Java getRelevantRules into JLanguageTool (no invent extras).
	lt := languagetool.NewJLanguageTool("fa")
	RegisterCorePersianRules(lt)
	// Order matches Java getRelevantRules registration order.
	require.Equal(t, language.PersianRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())

	// Core-rule outcomes (Java-visible without grammar.xml pattern pack):
	// Java demo text is commented out in PersianTest (config.ts DEMO_TEXTS noted but
	// alarms from pattern rules). Pattern IDs = honest gap without grammar.xml pack.

	// Multiple whitespace
	matches := lt.Check("این  متن  دو  فاصله.")
	ids := matchIDs(matches)
	require.Contains(t, ids, "WHITESPACE_RULE",
		"double spaces must match MultipleWhitespaceRule")

	// Double punctuation (Java DoublePunctuationRule: consecutive dots/commas only)
	matches = lt.Check("متن,, بد.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "DOUBLE_PUNCTUATION",
		",, must match DoublePunctuationRule")

	// PersianWordRepeatRule — Java example pair «این کار برای برای تو بود.»
	matches = lt.Check("این کار برای برای تو بود.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "PERSIAN_WORD_REPEAT_RULE",
		"برای برای must match PersianWordRepeatRule")

	// SimpleReplaceRule — official replace.txt pair (حاظر=حاضر, Java addExamplePair)
	matches = lt.Check("وی حاظر به همکاری شد.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "FA_SIMPLE_REPLACE",
		"حاظر must match SimpleReplaceRule (Java FA_SIMPLE_REPLACE)")

	// WordCoherencyRule — official coherency.txt pair اتاق/اطاق (WordCoherencyRuleTest)
	matches = lt.Check("این یک اتاق است. این یک اطاق است.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "FA_WORD_COHERENCY",
		"اتاق/اطاق must match WordCoherencyRule (Java FA_WORD_COHERENCY)")
}

// Analyze surface for Persian (runTests-adjacent) — sentence split only.
// Java runTests(lang, null, "ā"); pattern pack outcomes are honest gap.
func TestPersian_AnalyzeDemoText(t *testing.T) {
	// Java PersianTest has demo text commented out; use a simple multi-sentence sample.
	s := "این یک آزمایش است. لطفا متن خود را اینجا قرار دهید."
	lt := languagetool.NewJLanguageTool("fa")
	sents := lt.Analyze(s)
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
