package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PortugueseUnitConversionRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of PortugueseUnitConversionRuleTest.match (smoke + Java getId)
func TestPortugueseUnitConversionRule_Match(t *testing.T) {
	r := NewPortugueseUnitConversionRule(nil)
	require.Equal(t, "UNIDADES_METRICAS", r.GetID())
	require.NotNil(t, r)
	// Java: "A via tem 100 milhas de comprimento." expects metric suggestion
	sent := languagetool.AnalyzePlain("A via tem 100 milhas de comprimento.")
	matches := r.Match(sent)
	require.NotEmpty(t, matches)
}
