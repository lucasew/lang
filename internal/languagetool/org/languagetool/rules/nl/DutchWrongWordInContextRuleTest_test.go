package nl

// Twin of languagetool-language-modules/nl/src/test/java/org/languagetool/rules/nl/DutchWrongWordInContextRuleTest.java
// Java twin is @Ignore("no tests yet"); smoke-load dictionary.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDutchWrongWordInContextRule_Rule(t *testing.T) {
	rule := NewDutchWrongWordInContextRule(nil)
	require.Equal(t, "DUTCH_WRONG_WORD_IN_CONTEXT", rule.GetID())
	require.NotEmpty(t, rule.Entries)
}
