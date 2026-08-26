package languagetool

// Twin of languagetool-language-modules/pt JLanguageToolTest (module lang_pt).
// Full outcomes with RegisterCorePortugueseRules live in rules/pt/JLanguageToolTest_test.go
// (avoids languagetool↔rules/pt import cycles). Here: Analyze + package-local smoke.

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of JLanguageToolTest.testSomeSentences — no-panic smoke (package-local).
func TestJLanguageTool_lang_pt_SomeSentences(t *testing.T) {
	lt := NewJLanguageTool("pt-BR")
	require.Equal(t, "pt-BR", lt.GetLanguageCode())
	// trademark / multi-sentence smoke — Java only calls check
	_ = lt.Check("™ ® Marcas registradas da Corteva Agriscience e de suas companhias afiliadas.")
	_ = lt.Check("Zeth foi o mais novo de dez filhos.")
	// word-repeat via package-local inject (rules/pt has the real PortugueseWordRepeatRule)
	lt.AddRuleChecker("PORTUGUESE_WORD_REPEAT_RULE", SimpleWordRepeatChecker("PORTUGUESE_WORD_REPEAT_RULE"))
	require.Empty(t, lt.Check("Isto é uma frase. E outra."))
	require.NotEmpty(t, lt.Check("Isto é é uma frase."))
}

// Twin of JLanguageToolTest.testPortugueseVariants — structure smoke without invent.
// Full 0/1 agreement matrix needs grammar.xml; core-rule path is in rules/pt.
func TestJLanguageTool_lang_pt_PortugueseVariants(t *testing.T) {
	sentence := "Isto é uma característica sua."
	sentence2 := "Isto é uma características sua."
	for _, code := range []string{"pt-PT", "pt-BR", "pt-AO", "pt-MZ"} {
		lt := NewJLanguageTool(code)
		require.Equal(t, code, lt.GetLanguageCode())
		require.NotEmpty(t, lt.Analyze(sentence))
		// clean sentence empty without invent rules
		require.Empty(t, lt.Check(sentence))
		// agreement needs pattern rules — no invent match IDs
		_ = lt.Check(sentence2)
	}
}
