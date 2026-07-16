package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUnitConversionRuleUS_Match(t *testing.T) {
	rule := NewUnitConversionRuleUS(nil)
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("I just drank 3 pints."))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("I am 6 feet (2.02 m) tall."))))
	// range antipattern soft
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Use 3-5 pounds of butter."))))
}
