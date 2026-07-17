package patterns

// Twin of PatternRuleTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of PatternRuleTest.testSupportsLanguage
func TestPatternRule_SupportsLanguage(t *testing.T) {
	r := NewPatternRule("ID", "en", []*PatternToken{Token("foo")}, "d", "m", "s")
	require.True(t, r.SupportsLanguage("en"))
	require.True(t, r.SupportsLanguage("en-US"))
	require.True(t, r.SupportsLanguage("en-GB"))
	require.False(t, r.SupportsLanguage("de"))
	require.False(t, r.SupportsLanguage("de-DE"))

	de := NewPatternRule("D", "de-DE", nil, "d", "m", "s")
	require.True(t, de.SupportsLanguage("de"))
	require.True(t, de.SupportsLanguage("de-AT"))
	require.False(t, de.SupportsLanguage("en"))

	any := NewPatternRule("A", "", nil, "d", "m", "s")
	require.True(t, any.SupportsLanguage(""))
	require.False(t, any.SupportsLanguage("en")) // empty code only matches empty
}
