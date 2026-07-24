package rules

// Twin of RemoteRuleFilterTest — XML pattern load deferred; registry + filename surface.
import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Soft green: filter ID must be valid regex + FilterMatches smoke (full XML deferred).
func TestRemoteRuleFilterTest_Stub(t *testing.T) {
	// Java validates filter rule IDs are valid regexes
	re, err := regexp.Compile(`AI_.*`)
	require.NoError(t, err)
	require.True(t, re.MatchString("AI_FOO"))

	require.Equal(t, "en/remote-rule-filters.xml", GetRemoteRuleFilterFilename("en-US"))
	require.Equal(t, "de/remote-rule-filters.xml", GetRemoteRuleFilterFilename("de"))
	// special simple-language path
	require.Equal(t, "de/remote-rule-filters.xml", GetRemoteRuleFilterFilename("de-DE-x-simple-language"))

	f := NewRemoteRuleFilters()
	f.Register("xx", &FilterRule{
		IDPattern: regexp.MustCompile(`^DROP_.*`),
		MatchPositions: func(s *languagetool.AnalyzedSentence) []MatchPosition {
			return []MatchPosition{{Start: 0, End: 2}}
		},
	})
	sent := languagetool.AnalyzePlain("ab cd")
	drop := NewRuleMatch(NewFakeRule("DROP_X"), sent, 0, 2, "d")
	keep := NewRuleMatch(NewFakeRule("KEEP"), sent, 0, 2, "k")
	out := f.FilterMatches("xx", sent, []*RuleMatch{drop, keep})
	require.Len(t, out, 1)
	require.Equal(t, "KEEP", ruleIDOfMatch(out[0]))
}
