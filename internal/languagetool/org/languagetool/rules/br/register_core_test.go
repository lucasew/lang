package br

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("br")
	RegisterCoreBretonRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "BR_HA_HA")
}

// Java Breton.getRelevantRules exact ID set (layout subset + speller + topo).
func TestRegisterCoreBretonRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("br")
	RegisterCoreBretonRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"MORFOLOGIK_RULE_BR_FR",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"BR_TOPO",
	}
	require.ElementsMatch(t, want, ids)
	// No invent SharedLayout / word-repeat / compound extras
	for _, bad := range []string{
		"UNPAIRED_BRACKETS", "EMPTY_LINE", "TOO_LONG_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PUNCTUATION",
		"BR_COMPOUNDS", "WORD_REPEAT_RULE",
	} {
		require.NotContains(t, ids, bad)
	}
	for _, m := range lt.Check("test test") {
		require.NotContains(t, m.RuleID, "WORD_REPEAT")
	}
}
