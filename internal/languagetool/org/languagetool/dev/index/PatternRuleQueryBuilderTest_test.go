package index

// Twin of PatternRuleQueryBuilderTest — soft query string builder (Lucene deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Port of PatternRuleQueryBuilderTest (no @Test)
func TestPatternRuleQueryBuilder_NoTests(t *testing.T) {
	b := NewPatternRuleQueryBuilder("field")
	q := b.BuildSimple("hello", "world")
	require.Contains(t, q, `field:"hello"`)
	require.Contains(t, q, `field:"world"`)
	require.Contains(t, q, " AND ")

	toks := []*patterns.PatternToken{
		patterns.NewPatternTokenBuilder().Token("a").Build(),
		patterns.NewPatternTokenBuilder().Token("b").Build(),
	}
	require.Equal(t, `field:"a" AND field:"b"`, b.BuildFromTokens(toks))
	require.Equal(t, "*:*", b.BuildFromTokens(nil))
}
