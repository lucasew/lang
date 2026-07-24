package bitext

// Twin of languagetool-core/src/test/java/org/languagetool/rules/bitext/DifferentPunctuationRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of languagetool-core/src/test/java/org/languagetool/rules/bitext/DifferentPunctuationRuleTest.java :: DifferentPunctuationRuleTest.testRule
func TestDifferentPunctuationRule_Rule(t *testing.T) {
	r := NewDifferentPunctuationRule()
	require.Equal(t, "DIFFERENT_PUNCTUATION", r.GetID())
	require.NotEmpty(t, r.MatchBitext(multiWordSentence("Hi", "."), multiWordSentence("Hi", "!")))
	require.Empty(t, r.MatchBitext(multiWordSentence("Hi", "."), multiWordSentence("Hola", ".")))
	// non-punctuation last token → no flag
	require.Empty(t, r.MatchBitext(multiWordSentence("Hi"), multiWordSentence("Hi")))
}
