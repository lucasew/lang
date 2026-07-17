package patterns

// Twin of RuleNumberScalabilityTest — soft scale of PatternToken construction.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of RuleNumberScalabilityTest (no @Test)
func TestRuleNumberScalability_NoTests(t *testing.T) {
	const n = 100
	tokens := make([]*PatternToken, 0, n)
	for i := 0; i < n; i++ {
		tokens = append(tokens, NewPatternTokenBuilder().Token("w").Build())
	}
	require.Len(t, tokens, n)
	require.NotNil(t, tokens[0])
	require.NotNil(t, tokens[n-1])
}
