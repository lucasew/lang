package nl

// Twin of languagetool-language-modules/nl/src/test/java/org/languagetool/rules/nl/CompoundRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCompoundRule_Rule(t *testing.T) {
	rule := NewCompoundRule(nil)
	check := func(expectedErrors int, text string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(text))
		require.Equal(t, expectedErrors, len(matches), "text %q", text)
	}
	check(0, "Dit is een zee-egel.")
	check(0, "Zee-egel is een woord.")
	check(1, "Dit is een zee egel.")
	check(1, "Zee egel is een woord.")
}
