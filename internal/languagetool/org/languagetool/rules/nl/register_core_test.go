package nl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("nl")
	RegisterCoreDutchRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "NL_ALS_OF")
}

// Java Dutch.getRelevantRules exact ID set.
func TestRegisterCoreDutchRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("nl")
	RegisterCoreDutchRules(lt)
	require.ElementsMatch(t, language.DutchRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())
	for _, bad := range []string{
		"EMPTY_LINE", "WHITESPACE_PUNCTUATION", "WHITESPACE_PARAGRAPH",
		"WHITESPACE_PARAGRAPH_BEGIN", "PARAGRAPH_REPEAT_BEGINNING_RULE",
		"PUNCTUATION_PARAGRAPH_END", "WORD_REPEAT_RULE", "TOO_LONG_SENTENCE_NL",
		"NL_SENTENCE_WHITESPACE",
	} {
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), bad)
	}
}
