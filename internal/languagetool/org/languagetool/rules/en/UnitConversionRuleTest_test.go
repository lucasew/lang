package en

// Twin of UnitConversionRuleTest (simplified surface conversions).
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
	require.Equal(t, 1, matchN("The path is 100 miles long."))
	require.Equal(t, 0, matchN("I am 6 feet (1.82 m) tall."))
	require.Equal(t, 0, matchN("The path is 100 miles (160.93 km) long."))
}
