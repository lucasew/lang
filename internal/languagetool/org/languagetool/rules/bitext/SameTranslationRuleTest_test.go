package bitext

// Twin of languagetool-core/src/test/java/org/languagetool/rules/bitext/SameTranslationRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of languagetool-core/src/test/java/org/languagetool/rules/bitext/SameTranslationRuleTest.java :: SameTranslationRuleTest.testRule
func TestSameTranslationRule_Rule(t *testing.T) {
	r := NewSameTranslationRule()
	require.Equal(t, "SAME_TRANSLATION", r.GetID())
	// short (<4 nws tokens) same text → no flag
	require.Empty(t, r.MatchBitext(multiWordSentence("a", "b"), multiWordSentence("a", "b")))
	// long enough identical → flag
	words := []string{"This", "is", "same", "text"}
	require.NotEmpty(t, r.MatchBitext(multiWordSentence(words...), multiWordSentence(words...)))
	// different target → no flag
	require.Empty(t, r.MatchBitext(multiWordSentence(words...), multiWordSentence("other", "target", "text", "here")))
}
