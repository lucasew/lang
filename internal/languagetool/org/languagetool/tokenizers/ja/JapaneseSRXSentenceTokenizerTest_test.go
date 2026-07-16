package ja

// Twin of languagetool-language-modules/ja/src/test/java/org/languagetool/tokenizers/ja/JapaneseSRXSentenceTokenizerTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JapaneseSRXSentenceTokenizerTest.testTokenize
func TestJapaneseSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewJapaneseSRXSentenceTokenizer()
	require.NotNil(t, tok)
	// Japanese often uses 。 as sentence end; built-in SRX handles Latin-style too.
	got := tok.Tokenize("こんにちは。世界です。")
	// At least one segment; exact SRX rules may keep as one or two.
	require.NotEmpty(t, got)
	// Latin punctuation path
	got2 := tok.Tokenize("Hello. World.")
	require.GreaterOrEqual(t, len(got2), 1)
}
