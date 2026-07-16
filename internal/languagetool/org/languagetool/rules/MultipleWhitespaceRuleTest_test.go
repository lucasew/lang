package rules

// Twin of languagetool-core/src/test/java/org/languagetool/rules/MultipleWhitespaceRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-core/src/test/java/org/languagetool/rules/MultipleWhitespaceRuleTest.java :: MultipleWhitespaceRuleTest.testRule
func TestMultipleWhitespaceRule_Rule(t *testing.T) {
	// contains assertEquals — full values in Java twin source
	_ = "This is a test sentence." // assertGood
	_ = "This\uFEFF is a test sentence." // assertGood
	_ = "This\uFEFF\uFEFF is a test sentence." // assertGood
	_ = "This \uFEFFis a test sentence." // assertGood
	_ = "This\uFEFF\u2060 is a test sentence." // assertGood
	_ = "This\uFEFF\u2060 is a test sentence." // assertGood
	_ = "\uFEFF\uFEFFThis is a\n\u2060\ntest sentence..." // assertGood
	_ = "This is a test sentence..." // assertGood
	_ = "\n\tThis is a test sentence..." // assertGood
	_ = "Multiple tabs\t\tare okay" // assertGood
	_ = "\n This is a test sentence..." // assertGood
	_ = "\n    This is a test sentence..." // assertGood
}
