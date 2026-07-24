package jekavian

// Twin of languagetool-language-modules/sr/src/test/java/org/languagetool/rules/sr/jekavian/SimpleGrammarJekavianReplaceRuleTest.java
// Java twin has empty getMessage body; dictionary is currently header-only.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimpleGrammarJekavianReplaceRule_GetMessage(t *testing.T) {
	rule := NewSimpleGrammarJekavianReplaceRule(nil)
	require.Equal(t, "SR_JEKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE", rule.GetID())
}
