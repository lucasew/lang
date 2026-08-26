package patterns

// Twin of PatternRuleHandlerTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of PatternRuleHandlerTest.testReplaceSpacesInRegex
func TestPatternRuleHandler_ReplaceSpacesInRegex(t *testing.T) {
	s := "(?:[\\s\u00A0\u202F]+)"
	h := NewPatternRuleHandler("x", "en")
	require.Equal(t, "foo"+s+"bar", h.ReplaceSpacesInRegex("foo bar"))
	require.Equal(t, "foo"+s+"bar"+s+"x", h.ReplaceSpacesInRegex("foo bar x"))
	require.Equal(t, "foo"+s+s+"bar", h.ReplaceSpacesInRegex("foo  bar"))
	require.Equal(t, "fo[xy ]"+s+"bar", h.ReplaceSpacesInRegex("fo[xy ] bar"))
}
