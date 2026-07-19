package en

// Twin of languagetool-language-modules/en/.../UnitConversionRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUnitConversionRule_Match(t *testing.T) {
	rule := NewUnitConversionRule(nil)
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	require.Equal(t, 1, matchN("I am 6 feet tall."))
	// wrong parenthetical still flags (Java assertMatches … 1)
	require.Equal(t, 1, matchN("I am 6 feet (2.02 m) tall."))
	require.Equal(t, 0, matchN("I am 6 feet (1.82 m) tall."))
	require.Equal(t, 1, matchN("The path is 100 miles long."))
	require.Equal(t, 0, matchN("The path is 100 miles (160.93 km) long."))
	require.Equal(t, 1, matchN("It is 100 °F outside."))
	require.Equal(t, 1, matchN("My new apartment is 500 sq ft."))
}
