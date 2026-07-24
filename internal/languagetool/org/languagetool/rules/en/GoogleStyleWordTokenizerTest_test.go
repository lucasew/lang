package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/GoogleStyleWordTokenizerTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGoogleStyleWordTokenizer_Tokenize ports GoogleStyleWordTokenizerTest.testTokenize.
func TestGoogleStyleWordTokenizer_Tokenize(t *testing.T) {
	// Java: GoogleStyleWordTokenizer tokenizer = new GoogleStyleWordTokenizer();
	tokenizer := NewGoogleStyleWordTokenizer()
	// assertThat(tokenizer.tokenize(...), is(Arrays.asList(...)));
	require.Equal(t, []string{"foo", " ", "bar"}, tokenizer.Tokenize("foo bar"))
	require.Equal(t, []string{"foo", "-", "bar"}, tokenizer.Tokenize("foo-bar"))
	require.Equal(t, []string{"I", "'m", " ", "here", "."}, tokenizer.Tokenize("I'm here."))
	require.Equal(t, []string{"I", "'ll", " ", "do", " ", "that"}, tokenizer.Tokenize("I'll do that"))
	require.Equal(t, []string{"You", "'re", " ", "here"}, tokenizer.Tokenize("You're here"))
	require.Equal(t, []string{"You", "'ve", " ", "done", " ", "that"}, tokenizer.Tokenize("You've done that"))
}
