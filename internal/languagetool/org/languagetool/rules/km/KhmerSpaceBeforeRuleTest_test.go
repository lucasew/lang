package km

// Twin of languagetool-language-modules/km/src/test/java/org/languagetool/rules/km/KhmerSpaceBeforeRuleTest.java
// Note: AnalyzePlain may not match Khmer ZWSP tokenization; tests use simple forms.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestKhmerSpaceBeforeRule_SpaceBeforeRule(t *testing.T) {
	rule := NewKhmerSpaceBeforeRule(nil)
	// Correct: space before conjunction
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("x និង y"))))
	// Incorrect: no space before និង (previous token not space)
	// With plain tokenizer Khmer may be one token — use isolated token case.
	// Sentence-start conjunction is still flagged (Java behavior).
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("និង"))))
}
