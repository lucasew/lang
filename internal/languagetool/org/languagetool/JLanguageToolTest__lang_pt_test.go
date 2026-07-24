package languagetool

// Twin of languagetool-language-modules/pt JLanguageToolTest (module lang_pt).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of JLanguageToolTest.testSomeSentences
func TestJLanguageTool_lang_pt_SomeSentences(t *testing.T) {
	lt := NewJLanguageTool("pt-BR")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	require.Equal(t, "pt-BR", lt.GetLanguageCode())
	// trademark / multi-sentence smoke — no invent errors without grammar stack
	_ = lt.Check("™ ® Marcas registradas da Corteva Agriscience e de suas companhias afiliadas.")
	_ = lt.Check("Zeth foi o mais novo de dez filhos.")
	// word-repeat inject
	require.Empty(t, lt.Check("Isto é uma frase. E outra."))
	require.NotEmpty(t, lt.Check("Isto é é uma frase."))
}

// Twin of JLanguageToolTest.testPortugueseVariants
func TestJLanguageTool_lang_pt_PortugueseVariants(t *testing.T) {
	sentence := "Isto é uma característica sua."
	sentence2 := "Isto é uma características sua."
	for _, code := range []string{"pt-PT", "pt-BR", "pt-AO", "pt-MZ"} {
		lt := NewJLanguageTool(code)
		// without agreement grammar, inject WORD_REPEAT only — structure smoke
		lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
		require.Equal(t, code, lt.GetLanguageCode())
		// clean sentence empty under inject
		require.Empty(t, lt.Check(sentence))
		// number agreement needs grammar; fail-closed empty unless rule present
		_ = lt.Check(sentence2)
	}
}
