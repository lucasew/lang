package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/GoogleStyleWordTokenizerTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoogleStyleWordTokenizer_Tokenize(t *testing.T) {
	w := NewGoogleStyleWordTokenizer()
	require.Equal(t, []string{"foo", " ", "bar"}, w.Tokenize("foo bar"))
	require.Equal(t, []string{"foo", "-", "bar"}, w.Tokenize("foo-bar"))
	require.Equal(t, []string{"I", "'m", " ", "here", "."}, w.Tokenize("I'm here."))
	require.Equal(t, []string{"I", "'ll", " ", "do", " ", "that"}, w.Tokenize("I'll do that"))
	require.Equal(t, []string{"You", "'re", " ", "here"}, w.Tokenize("You're here"))
	require.Equal(t, []string{"You", "'ve", " ", "done", " ", "that"}, w.Tokenize("You've done that"))
}
