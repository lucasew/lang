package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// MockFilter ports the Java test helper MockFilter (no-arg ctor RuleFilter).
type MockFilter struct{}

func (MockFilter) AcceptRuleMatch(match *rules.RuleMatch, _ map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	return match
}

func init() {
	// Register under Java FQCN so getFilter(MockFilter.class.getName()) twin works.
	GlobalRuleFilterCreator.Register("org.languagetool.rules.patterns.MockFilter", func() RuleFilter {
		return MockFilter{}
	})
}

// Twin of RuleFilterCreatorTest.testMockFilter
func TestRuleFilterCreator_MockFilter(t *testing.T) {
	// Java: RuleFilterCreator.getInstance().getFilter(MockFilter.class.getName())
	filter := GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.patterns.MockFilter")
	require.NotNil(t, filter)
}

// Twin of RuleFilterCreatorTest.testInvalidClassName
func TestRuleFilterCreator_InvalidClassName(t *testing.T) {
	// Java: getFilter("MyInvalidClassName") → RuntimeException
	require.Panics(t, func() {
		GlobalRuleFilterCreator.GetFilter("MyInvalidClassName")
	})
}
