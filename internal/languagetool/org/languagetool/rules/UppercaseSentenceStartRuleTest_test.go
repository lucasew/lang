package rules

// Twin of languagetool-core/src/test/java/org/languagetool/rules/UppercaseSentenceStartRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-core/src/test/java/org/languagetool/rules/UppercaseSentenceStartRuleTest.java :: UppercaseSentenceStartRuleTest.testRule
func TestUppercaseSentenceStartRule_Rule(t *testing.T) {
	// contains assertEquals — full values in Java twin source
	_ = "this" // assertGood
	_ = "a) This is a test sentence." // assertGood
	_ = "iv. This is a test sentence..." // assertGood
	_ = "\"iv. This is a test sentence...\"" // assertGood
	_ = "»iv. This is a test sentence..." // assertGood
	_ = "This" // assertGood
	_ = "This is" // assertGood
	_ = "This is a test sentence" // assertGood
	_ = "" // assertGood
	_ = "http: 
    assertGood(" // assertGood
	_ = "¿Esto es una pregunta?" // assertGood
	_ = "¿Esto es una pregunta?, ¿y esto?" // assertGood
	_ = "ø This is a test sentence with a wrong bullet character." // assertGood
}
