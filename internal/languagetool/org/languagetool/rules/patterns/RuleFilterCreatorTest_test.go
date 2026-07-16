package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

type noopFilter2 struct{}

func (noopFilter2) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	return match
}

func TestRuleFilterCreator_TestInvalidClassName(t *testing.T) {
	// invalid / unregistered class name panics (Java ClassNotFound-like)
	c := NewRuleFilterCreator()
	require.Panics(t, func() { c.GetFilter("org.languagetool.rules.DoesNotExist") })
	// valid registration works
	c.Register("org.example.Noop", func() RuleFilter { return noopFilter2{} })
	require.NotNil(t, c.GetFilter("org.example.Noop"))
}
