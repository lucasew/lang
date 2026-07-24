package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("es")
	RegisterCoreSpanishRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "ES_A_EL")
}

// Java Spanish.getRelevantRules exact ID set.
func TestRegisterCoreSpanishRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("es")
	RegisterCoreSpanishRules(lt)
	require.ElementsMatch(t, language.SpanishRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())
	for _, bad := range []string{
		"EMPTY_LINE", "SENTENCE_WHITESPACE", "WHITESPACE_PUNCTUATION",
		"WHITESPACE_PARAGRAPH", "UNPAIRED_BRACKETS", "TOO_LONG_SENTENCE_ES",
		"ES_SENTENCE_WHITESPACE",
	} {
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), bad)
	}
}
