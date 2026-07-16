package ca

// Twin of ReplaceOperationNamesRuleTest — surface dictionary (no POS filters).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestReplaceOperationNamesRule_Rule(t *testing.T) {
	rule := NewReplaceOperationNamesRule(nil)
	// incorrect (Java twin)
	for _, s := range []string{
		"Assecat del braç del riu",
		"Cal vigilar el filtrat del vi",
		"El procés d'empaquetat",
	} {
		matches := rule.Match(languagetool.AnalyzePlain(s))
		require.NotEmpty(t, matches, s)
	}
}
