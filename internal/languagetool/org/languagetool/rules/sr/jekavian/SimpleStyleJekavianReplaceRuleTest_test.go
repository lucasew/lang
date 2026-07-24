package jekavian

// Twin of languagetool-language-modules/sr/src/test/java/org/languagetool/rules/sr/jekavian/SimpleStyleJekavianReplaceRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimpleStyleJekavianReplaceRule_GetMessage(t *testing.T) {
	rule := NewSimpleStyleJekavianReplaceRule(nil)
	require.Equal(t, "SR_JEKAVIAN_SIMPLE_STYLE_REPLACE_RULE", rule.GetID())
}
