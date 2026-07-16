package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUnitConversionRuleImperial_Match(t *testing.T) {
	rule := NewUnitConversionRuleImperial(nil)
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("I just drank 3 pints."))))
	// wrong metres in paren still flags feet
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("I am 6 feet (2.02 m) tall."))))
}
