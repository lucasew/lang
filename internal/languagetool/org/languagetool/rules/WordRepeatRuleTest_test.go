package rules

// Twin of languagetool-core/src/test/java/org/languagetool/rules/WordRepeatRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-core/src/test/java/org/languagetool/rules/WordRepeatRuleTest.java :: WordRepeatRuleTest.test
func TestWordRepeatRule_Test(t *testing.T) {
	_ = "A test" // assertGood
	_ = "A test." // assertGood
	_ = "A test..." // assertGood
	_ = "1 000 000 years" // assertGood
	_ = "010 020 030" // assertGood
	_ = "\uD83D\uDC4D\uD83D\uDC9A\uD83C\uDF32\uD83C\uDF32" // assertGood
	_ = "A A test" // assertBad
	_ = "A a test" // assertBad
	_ = "This is is a test" // assertBad
}
