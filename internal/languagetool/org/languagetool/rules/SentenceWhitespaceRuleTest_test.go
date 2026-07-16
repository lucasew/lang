package rules

// Twin of languagetool-core/src/test/java/org/languagetool/rules/SentenceWhitespaceRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-core/src/test/java/org/languagetool/rules/SentenceWhitespaceRuleTest.java :: SentenceWhitespaceRuleTest.testMatch
func TestSentenceWhitespaceRule_Match(t *testing.T) {
	_ = "This is a text. And there's the next sentence." // assertGood
	_ = "This is a text! And there's the next sentence." // assertGood
	_ = "This is a text\nAnd there's the next sentence." // assertGood
	_ = "This is a text\n\nAnd there's the next sentence." // assertGood
	_ = "This is a text.And there's the next sentence." // assertBad
	_ = "This is a text!And there's the next sentence." // assertBad
	_ = "This is a text?And there's the next sentence." // assertBad
}
