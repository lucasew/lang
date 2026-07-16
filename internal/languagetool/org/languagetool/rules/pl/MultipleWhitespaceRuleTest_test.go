package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/rules/pl/MultipleWhitespaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestMultipleWhitespaceRule_Rule(t *testing.T) {
	rule := rules.NewMultipleWhitespaceRule(nil)
	require.Equal(t, 0, len(rule.Match([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("To jest test.")})))
	require.Equal(t, 1, len(rule.Match([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("To jest   test.")})))
}
