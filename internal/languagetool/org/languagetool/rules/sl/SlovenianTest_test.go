package sl

// Twin of languagetool-language-modules/sl/src/test/java/org/languagetool/rules/sl/SlovenianTest.java
//
// Java testLanguage:
//   demo text expects: TROPIČJETRI, PREDLOG_S, KI_BREZ_VEJICE, MORFOLOGIK_RULE_SL_SI,
//   COMMA_PARENTHESIS_WHITESPACE, WORD_REPEAT_RULE
//   + runTests(lang)
//
// Pattern IDs TROPIČJETRI / PREDLOG_S / KI_BREZ_VEJICE come from grammar.xml and are
// not auto-loaded by NewJLanguageTool("sl") + RegisterCoreSlovenianRules without
// activateDefaultPatternRules. This twin asserts Java-visible surface the port can
// prove without invent: metadata, getRelevantRules IDs, and real core-rule outcomes
// (comma/paren whitespace, word repeat, speller when dict loads). Honest gap:
// grammar.xml pattern matches not asserted (same class as RO/SK).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of SlovenianTest.testLanguage — metadata + relevant rules + provable core outcomes.
func TestSlovenian_Language(t *testing.T) {
	lang := language.NewSlovenian()
	require.Equal(t, "sl", lang.GetShortCode())
	require.Equal(t, "Slovenian", lang.GetName())
	require.Equal(t, []string{"SI"}, lang.GetCountries())
	require.Equal(t, languagetool.LookingForNewMaintainer, lang.GetMaintainedState())
	require.Equal(t, "Martin Srebotnjak", lang.GetMaintainers()[0].Name)
	require.Equal(t, "MORFOLOGIK_RULE_SL_SI", lang.GetDefaultSpellingRuleID())

	// getRelevantRules IDs in Java Arrays.asList order
	require.Equal(t, language.SlovenianRelevantRuleIDs(), lang.GetRelevantRuleIDs())
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"MORFOLOGIK_RULE_SL_SI",
		"UPPERCASE_SENTENCE_START",
		"WORD_REPEAT_RULE",
		"WHITESPACE_RULE",
	}, lang.GetRelevantRuleIDs())

	// Wire Java getRelevantRules into JLanguageTool (no invent extras).
	lt := languagetool.NewJLanguageTool("sl")
	RegisterCoreSlovenianRules(lt)
	// Order matches Java getRelevantRules registration order.
	require.Equal(t, language.SlovenianRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())

	// Core-rule outcomes (Java-visible without grammar.xml pattern pack):
	// Java demo text (config.ts DEMO_TEXTS / SlovenianTest):
	//   "Tukaj vnesite svoje besedilo... Pa poglejmo primer besedila s nekaj napakami
	//    ki jih lahko razpozna orodje LanguageTool; ko opazite napake, jih lahko
	//    enostavno popiravite. ( Obenem se izvrši tudi preverjanje črkovanja črkovanja."
	// Core IDs from that demo: MORFOLOGIK_RULE_SL_SI, COMMA_PARENTHESIS_WHITESPACE,
	// WORD_REPEAT_RULE. Pattern IDs TROPIČJETRI/PREDLOG_S/KI_BREZ_VEJICE = honest gap.

	// COMMA_PARENTHESIS_WHITESPACE — space after opening paren (demo "( Obenem")
	matches := lt.Check("Besedilo. ( Obenem se izvrši.")
	ids := matchIDs(matches)
	require.Contains(t, ids, "COMMA_PARENTHESIS_WHITESPACE",
		"'( ' must match CommaWhitespaceRule (Java demo COMMA_PARENTHESIS_WHITESPACE)")

	// WORD_REPEAT_RULE — demo "črkovanja črkovanja"
	matches = lt.Check("Preverjanje črkovanja črkovanja.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "WORD_REPEAT_RULE",
		"'črkovanja črkovanja' must match WordRepeatRule")

	// Unpaired brackets (Java Slovenian GenericUnpairedBracketsRule symbols)
	matches = lt.Check("To je besedilo „brez zaprtja.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "UNPAIRED_BRACKETS",
		"unpaired „ must match GenericUnpairedBracketsRule (Java UNPAIRED_BRACKETS)")

	// Multiple whitespace
	matches = lt.Check("Tu  sta  dve presledki.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "WHITESPACE_RULE",
		"double spaces must match MultipleWhitespaceRule")

	// Double punctuation (Java DoublePunctuationRule: consecutive dots/commas only)
	matches = lt.Check("Besedilo,, napačno.")
	ids = matchIDs(matches)
	require.Contains(t, ids, "DOUBLE_PUNCTUATION",
		",, must match DoublePunctuationRule")

	// Speller — require ≥1 hit when official sl_SI.dict loads
	// Java demo uses "popiravite" (typo) → MORFOLOGIK_RULE_SL_SI
	if morfologik.DiscoverLanguageDict(SlovenianSpellerDict) == "" {
		t.Log("sl_SI.dict not discoverable — speller outcome skipped (honest gap for missing resource)")
	} else {
		matches = lt.Check("Lahko enostavno popiravite napake.")
		ids = matchIDs(matches)
		spellerHits := 0
		for _, id := range ids {
			if id == "MORFOLOGIK_RULE_SL_SI" {
				spellerHits++
			}
		}
		require.GreaterOrEqual(t, spellerHits, 1,
			"deliberate misspelling must hit MorfologikSlovenianSpellerRule when sl_SI.dict loads")
	}
}

// Analyze surface for Java demo text (runTests-adjacent) — sentence split only.
// Pattern rule IDs from full testDemoText are honest gap without grammar.xml pack.
func TestSlovenian_AnalyzeDemoText(t *testing.T) {
	// Java SlovenianTest demo string (config.ts DEMO_TEXTS)
	s := "Tukaj vnesite svoje besedilo... Pa poglejmo primer besedila s nekaj napakami ki jih lahko razpozna orodje LanguageTool; ko opazite napake, jih lahko enostavno popiravite. ( Obenem se izvrši tudi preverjanje črkovanja črkovanja."
	lt := languagetool.NewJLanguageTool("sl")
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
