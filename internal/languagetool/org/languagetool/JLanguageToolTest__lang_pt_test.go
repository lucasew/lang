package languagetool

// Twin of JLanguageToolTest for Portuguese — analysis smoke.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJLanguageTool_SomeSentences(t *testing.T) {
	lt := NewJLanguageTool("pt")
	require.Equal(t, "pt", lt.GetLanguageCode())
	sents := lt.Analyze("Isto é uma frase. E outra.")
	require.NotEmpty(t, sents)
}
