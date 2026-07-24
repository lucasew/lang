package fa

// Twin of languagetool-language-modules/fa/src/test/java/org/languagetool/rules/fa/PersianSpaceBeforeRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPersianSpaceBeforeRule_Rules(t *testing.T) {
	rule := NewPersianSpaceBeforeRule(nil)
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("به اینجا"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("من به اینجا"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("(به اینجا"))))
}
