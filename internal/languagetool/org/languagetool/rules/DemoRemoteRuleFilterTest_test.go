package rules

// Twin of DemoRemoteRuleFilterTest — demo language filter path smoke.
import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of DemoRemoteRuleFilterTest.testRules (full grammar XML deferred).
func TestDemoRemoteRuleFilter_Rules(t *testing.T) {
	// demo language short code is typically "xx"
	require.Equal(t, "xx/remote-rule-filters.xml", GetRemoteRuleFilterFilename("xx"))

	f := NewRemoteRuleFilters()
	f.Register("xx", &FilterRule{
		IDPattern: regexp.MustCompile(`AI_.*`),
		MatchPositions: func(s *languagetool.AnalyzedSentence) []MatchPosition {
			if s == nil {
				return nil
			}
			// filter whole sentence span soft
			return []MatchPosition{{Start: 0, End: len(s.GetText())}}
		},
	})
	sent := languagetool.AnalyzePlain("demo text")
	drop := NewRuleMatch(NewFakeRule("AI_DEMO"), sent, 0, len(sent.GetText()), "drop")
	keep := NewRuleMatch(NewFakeRule("GRAMMAR"), sent, 0, 4, "keep")
	out := f.FilterMatches("xx", sent, []*RuleMatch{drop, keep})
	require.Len(t, out, 1)
	require.Equal(t, "GRAMMAR", ruleIDOfMatch(out[0]))
}
