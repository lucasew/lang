package rules

// Twin of languagetool-core/src/test/java/org/languagetool/rules/AbstractCompoundRuleTest.java
// Abstract base class in Java; concrete coverage lives in language CompoundRule tests (e.g. en.CompoundRuleTest).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAbstractCompoundRuleTest_Stub(t *testing.T) {
	// No standalone tests — see language-module CompoundRuleTest twins.
}

// Java AbstractCompoundRule.normalize: String.trim + Pattern("\\s+") ASCII-only.
func TestNormalizeCompound_JavaTrimAndWhitespace(t *testing.T) {
	require.Equal(t, "foo bar", normalizeCompound("  foo   bar  "))
	require.Equal(t, "foo bar", normalizeCompound("foo - bar"))
	require.Equal(t, "foo bar", normalizeCompound("foo-bar"))
	// NBSP is not stripped by String.trim and not matched by Java \\s
	require.Equal(t, "\u00a0foo", normalizeCompound("\u00a0foo"))
	// multiple ASCII spaces collapse
	require.Equal(t, "a b", normalizeCompound("a\t\nb"))
}
