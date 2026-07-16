package rules

// Twin of RemoteRuleFiltersTest
import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRemoteRuleFilters_Load(t *testing.T) {
	// XML load deferred; filename path surface + empty registry load shape.
	require.Equal(t, "xx/remote-rule-filters.xml", GetRemoteRuleFilterFilename("xx"))
	require.Equal(t, "de/remote-rule-filters.xml", GetRemoteRuleFilterFilename("de-DE-x-simple-language"))
	f := NewRemoteRuleFilters()
	require.Empty(t, f.FilterMatches("xx", languagetool.AnalyzePlain("a"), nil))
}

func TestRemoteRuleFilters_SimpleFilter(t *testing.T) {
	f := NewRemoteRuleFilters()
	f.Register("en", &FilterRule{
		IDPattern: regexp.MustCompile(`^TEST_REMOTE_RULE$`),
		MatchPositions: func(s *languagetool.AnalyzedSentence) []MatchPosition {
			return []MatchPosition{{Start: 0, End: 4}}
		},
	})
	sent := languagetool.AnalyzePlain("test sentence")
	drop := NewRuleMatch(NewFakeRule("TEST_REMOTE_RULE"), sent, 0, 4, "y")
	keep := NewRuleMatch(NewFakeRule("OTHER"), sent, 0, 4, "x")
	out := f.FilterMatches("en", sent, []*RuleMatch{drop, keep})
	require.Len(t, out, 1)
	require.Equal(t, "OTHER", ruleIDOfMatch(out[0]))
}

func TestRemoteRuleFilters_MultiTokenWhitespace(t *testing.T) {
	f := NewRemoteRuleFilters()
	// Filter spans across multi-token whitespace region 0-11 "test sentence"
	f.Register("en", &FilterRule{
		IDPattern: regexp.MustCompile(`AI_.*`),
		MatchPositions: func(s *languagetool.AnalyzedSentence) []MatchPosition {
			return []MatchPosition{{Start: 0, End: 13}}
		},
	})
	sent := languagetool.AnalyzePlain("test sentence")
	drop := NewRuleMatch(NewFakeRule("AI_WS"), sent, 0, 13, "y")
	out := f.FilterMatches("en", sent, []*RuleMatch{drop})
	require.Empty(t, out)
}

func TestRemoteRuleFilters_Marker(t *testing.T) {
	// Marker-like filter: empty positions → drop all matching IDs
	f := NewRemoteRuleFilters()
	f.Register("en", &FilterRule{
		IDPattern: regexp.MustCompile(`MARK_.*`),
	})
	sent := languagetool.AnalyzePlain("abc")
	drop := NewRuleMatch(NewFakeRule("MARK_1"), sent, 0, 1, "m")
	keep := NewRuleMatch(NewFakeRule("OK"), sent, 0, 1, "k")
	out := f.FilterMatches("en", sent, []*RuleMatch{drop, keep})
	require.Len(t, out, 1)
	require.Equal(t, "OK", ruleIDOfMatch(out[0]))
}

func TestRemoteRuleFilters_Antipattern(t *testing.T) {
	// Antipattern surface: non-overlapping span keeps match
	f := NewRemoteRuleFilters()
	f.Register("en", &FilterRule{
		IDPattern: regexp.MustCompile(`AI_.*`),
		MatchPositions: func(s *languagetool.AnalyzedSentence) []MatchPosition {
			return []MatchPosition{{Start: 0, End: 2}}
		},
	})
	sent := languagetool.AnalyzePlain("abcdef")
	keep := NewRuleMatch(NewFakeRule("AI_X"), sent, 3, 5, "y")
	out := f.FilterMatches("en", sent, []*RuleMatch{keep})
	require.Len(t, out, 1)
}

func TestRemoteRuleFilters_IDRegexFilter(t *testing.T) {
	f := NewRemoteRuleFilters()
	f.Register("en-US", &FilterRule{
		IDPattern: regexp.MustCompile(`^GPT_.*_(SUGGEST|FIX)$`),
	})
	sent := languagetool.AnalyzePlain("hi")
	drop := NewRuleMatch(NewFakeRule("GPT_EN_SUGGEST"), sent, 0, 2, "d")
	keep := NewRuleMatch(NewFakeRule("GPT_EN_OTHER"), sent, 0, 2, "k")
	// short-code fallback from en-US registration when looking up en-US
	out := f.FilterMatches("en-US", sent, []*RuleMatch{drop, keep})
	require.Len(t, out, 1)
	require.Equal(t, "GPT_EN_OTHER", ruleIDOfMatch(out[0]))
}
