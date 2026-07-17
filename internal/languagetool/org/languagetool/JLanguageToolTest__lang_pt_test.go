package languagetool

// Twin of PT JLanguageToolTest — Check inject.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJLanguageTool_SomeSentences(t *testing.T) {
	lt := NewJLanguageTool("pt")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	require.Equal(t, "pt", lt.GetLanguageCode())
	require.Empty(t, lt.Check("Isto é uma frase. E outra."))
	require.NotEmpty(t, lt.Check("Isto é é uma frase."))
}
