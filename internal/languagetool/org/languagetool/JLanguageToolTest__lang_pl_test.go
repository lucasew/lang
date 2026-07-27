package languagetool

// Twin stub colocated with other JLanguageToolTest__lang_* files.
// Full testPolish outcomes (RegisterCorePolishRules + enableRule) live in
// rules/pl/JLanguageToolTest_test.go to avoid languagetool↔rules/pl import cycles
// in same-package tests. Here: Analyze + WordRepeat smoke with no invent soft path.

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testPolish (package-local subset without rules/pl import).
func TestJLanguageTool_lang_pl_Polish(t *testing.T) {
	lt := NewJLanguageTool("pl")
	require.Equal(t, "pl", lt.GetLanguageCode())
	// Correct sentence analyzes
	require.NotEmpty(t, lt.Analyze("To jest całkowicie prawidłowe zdanie."))
	// Word repeat via real WordRepeatRule checker (SimpleWordRepeatChecker is the
	// package-local twin of rules.WordRepeatRule used when rules/pl cannot be imported).
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	require.Empty(t, lt.Check("To jest całkowicie prawidłowe zdanie."))
	matches := lt.Check("To jest jest problem.")
	require.Len(t, matches, 1)
	require.Equal(t, "WORD_REPEAT_RULE", matches[0].RuleID)
}
