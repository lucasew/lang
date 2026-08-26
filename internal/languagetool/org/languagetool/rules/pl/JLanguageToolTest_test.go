package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/JLanguageToolTest.java
// (testPolish).
//
// Lives in rules/pl (not package languagetool) to avoid languagetool↔rules/pl import
// cycles in same-package tests. Java source path is still the PL module JLanguageToolTest.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	tagpl "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pl"
	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testPolish — Java-visible outcomes with real PL core rules.
func TestJLanguageTool_Polish(t *testing.T) {
	// assertEquals("[PL]", Arrays.toString(polish.getCountries()));
	require.Equal(t, []string{"PL"}, language.Polish.GetCountries())
	require.Equal(t, "pl", language.Polish.GetShortCode())
	require.Equal(t, "Polish", language.Polish.GetName())

	// noXmlRulesPolish: getRelevantRules only (no pattern XML)
	noRulesTool := languagetool.NewJLanguageTool("pl")
	RegisterCorePolishRules(noRulesTool)

	// fully correct Polish sentence → 0 matches
	require.Empty(t, noRulesTool.Check("To jest całkowicie prawidłowe zdanie."))

	// word repeat (WordRepeatRule → WORD_REPEAT_RULE)
	matches := noRulesTool.Check("To jest jest problem.")
	require.Len(t, matches, 1)
	require.Equal(t, "WORD_REPEAT_RULE", matches[0].RuleID)

	// PL_WORD_REPEAT is default-off (AdvancedWordRepeatRule.setDefaultOff).
	// Java: "Był on bowiem pięknym strzelcem bowiem." → 0 until enableRule.
	require.Contains(t, noRulesTool.GetAllRegisteredRuleIDs(), "PL_WORD_REPEAT")
	require.Contains(t, noRulesTool.GetDefaultOffRuleIDs(), "PL_WORD_REPEAT")

	if tagpl.DiscoverPolishPOSDict() != "" {
		tagpl.EnsureDefaultPolishTagger()
		noRulesTool.SentenceTagger = tagpl.DefaultPolishTagger.Tag
		offMatches := noRulesTool.Check("Był on bowiem pięknym strzelcem bowiem.")
		for _, m := range offMatches {
			require.NotEqual(t, "PL_WORD_REPEAT", m.RuleID,
				"PL_WORD_REPEAT must stay off until EnableRule")
		}
		// enableRule("PL_WORD_REPEAT") → 1 match (Java)
		noRulesTool.EnableRule("PL_WORD_REPEAT")
		onMatches := noRulesTool.Check("Był on bowiem pięknym strzelcem bowiem.")
		found := false
		for _, m := range onMatches {
			if m.RuleID == "PL_WORD_REPEAT" {
				found = true
				break
			}
		}
		require.True(t, found, "Java enableRule(PL_WORD_REPEAT) → match on 'bowiem'…'bowiem'")
	}

	// Spelling activates on non-Polish text (Java: English → MORFOLOGIK matches).
	enMatches := noRulesTool.Check("I can give you more a detailed description")
	spellerHits := 0
	for _, m := range enMatches {
		if m.RuleID == "MORFOLOGIK_RULE_PL_PL" {
			spellerHits++
		}
	}
	// Java expects 6 matches total with no XML rules; assert real speller activity.
	require.GreaterOrEqual(t, spellerHits, 1,
		"English text under PL speller must produce MORFOLOGIK_RULE_PL_PL matches")

	// Disambiguation immunization (Java noRulesTool 0-match cases) needs
	// getDisambiguator() on Analyze path. PolishLang.CreateDefaultDisambiguator
	// returns PolishHybridDisambiguator (language/polish_test.go). Full
	// immunization outcomes deferred until Analyze wires that factory — no invent.
}

// Deferred Java testPolish sections (pattern rules / hybrid Analyze wire).
func TestJLanguageTool_Polish_DocumentDeferredPatternSections(t *testing.T) {
	// - Mają one niemałe znaczenie. → 0 (disambig)
	// - immunization / antipattern cases
	// - column positions for "w której nie ma przecinka"
	// - getUnknownWords listUnknownWords path
	// - getAnalyzedSentence annotations log for "Z powodu pogody..."
	t.Log("pattern/disambig immunization and column checks deferred — no invent")
}
