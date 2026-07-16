package el

// Twin of languagetool-language-modules/el/src/test/java/org/languagetool/rules/el/NumeralStressRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestNumeralStressRule_Rule(t *testing.T) {
	rule := NewNumeralStressRule(nil)
	ok := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(s))), s)
	}
	bad := func(s, want string) {
		t.Helper()
		m := rule.Match(languagetool.AnalyzePlain(s))
		require.Equal(t, 1, len(m), s)
		require.Equal(t, want, m[0].GetSuggestedReplacements()[0], s)
	}
	ok("1ος")
	ok("2η")
	ok("3ο")
	ok("20ός")
	ok("30ή")
	ok("40ό")
	ok("1000ών")
	ok("1010ες")

	bad("4ός", "4ος")
	bad("5ή", "5η")
	bad("6ό", "6ο")
	bad("100ος", "100ός")
	bad("200η", "200ή")
	bad("300ο", "300ό")
	bad("2000ων", "2000ών")
	bad("2010ές", "2010ες")
	bad("2020α", "2020ά")
}
