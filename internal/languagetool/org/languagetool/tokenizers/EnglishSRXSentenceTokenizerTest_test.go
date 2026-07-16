package tokenizers

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/tokenizers/EnglishSRXSentenceTokenizerTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnglishSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewEnglishSRXSentenceTokenizer()
	require.NotNil(t, tok)
	got := tok.Tokenize("This is a sentence. And another one.")
	require.GreaterOrEqual(t, len(got), 1)
	require.NotEmpty(t, got[0])
}
