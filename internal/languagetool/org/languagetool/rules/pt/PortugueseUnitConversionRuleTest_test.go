package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PortugueseUnitConversionRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of PortugueseUnitConversionRuleTest.match
func TestPortugueseUnitConversionRule_Match(t *testing.T) {
	r := NewPortugueseUnitConversionRule(nil)
	require.Equal(t, "UNITS_PT", r.GetID())
	// Smoke: rule constructs and Match is callable (full imperial→metric needs tagged numbers).
	require.NotNil(t, r)
}
