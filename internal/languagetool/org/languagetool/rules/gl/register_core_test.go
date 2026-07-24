package gl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("gl")
	RegisterCoreGalicianRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "GL_A_O")
}

// Java Galician.getRelevantRules exact ID set.
func TestRegisterCoreGalicianRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("gl")
	RegisterCoreGalicianRules(lt)
	require.ElementsMatch(t, language.GalicianRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())
	for _, bad := range []string{
		"GL_UNPAIRED_BRACKETS", "TOO_LONG_SENTENCE_GL", "WORD_REPEAT_RULE",
		"WHITESPACE_PUNCTUATION",
	} {
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), bad)
	}
}
