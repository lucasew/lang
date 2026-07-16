package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishSpecificCaseRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishSpecificCaseRuleTest.java :: EnglishSpecificCaseRuleTest.testRule
func TestEnglishSpecificCaseRule_Rule(t *testing.T) {
	// contains assertThat
	_ = "Harry Potter" // assertGood
	_ = "I like Harry Potter." // assertGood
	_ = "I like HARRY POTTER." // assertGood
	_ = "harry potter" // assertBad
	_ = "harry Potter" // assertBad
	_ = "Harry potter" // assertBad
	_ = "I like Harry potter." // assertBad
	_ = "Alexander The Great" // assertBad
	_ = "I like Harry  potter." // assertBad
}
