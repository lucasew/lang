package bitext

// Twin of languagetool-core/src/test/java/org/languagetool/rules/bitext/DifferentLengthRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of languagetool-core/src/test/java/org/languagetool/rules/bitext/DifferentLengthRuleTest.java :: DifferentLengthRuleTest.testRule
func TestDifferentLengthRule_Rule(t *testing.T) {
	r := NewDifferentLengthRule()
	require.Equal(t, "TRANSLATION_LENGTH", r.GetID())
	// very short target vs long source
	require.NotEmpty(t, r.MatchBitext(sentence("this is a longer source sentence"), sentence("x")))
	// similar lengths
	require.Empty(t, r.MatchBitext(sentence("hello world"), sentence("hola mundo!")))
}
