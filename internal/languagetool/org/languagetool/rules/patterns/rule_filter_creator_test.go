package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

type noopFilter struct{}

func (noopFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	return match
}

func TestRuleFilterCreatorPort(t *testing.T) {
	c := NewRuleFilterCreator()
	c.Register("org.example.Noop", func() RuleFilter { return noopFilter{} })
	f := c.GetFilter("org.example.Noop")
	require.NotNil(t, f)
	// cached
	require.Equal(t, f, c.GetFilter("org.example.Noop"))
	require.Panics(t, func() { c.GetFilter("missing.Filter") })
}
