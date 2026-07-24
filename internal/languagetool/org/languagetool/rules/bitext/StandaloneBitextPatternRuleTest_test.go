package bitext

// Twin of StandaloneBitextPatternRuleTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStandaloneBitextPatternRule_BitextPatternRuleTest(t *testing.T) {
	// Built-in structural bitext rules always available.
	rules := RelevantBitextRules()
	require.NotEmpty(t, rules)
	// SameTranslation needs >3 non-whitespace tokens and identical text.
	matches := CheckBitext("This is a test sentence here", "This is a test sentence here", nil)
	require.NotEmpty(t, matches)
}
