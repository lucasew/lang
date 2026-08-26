package patterns

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func xxRemoteFiltersPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// .../internal/languagetool/org/languagetool/rules/patterns → repo root (6 levels up)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "..", ".."))
	p := filepath.Join(root, "testdata/upstream/xx/remote-rule-filters.xml")
	require.FileExists(t, p)
	return p
}

func TestLoadRemoteRuleFiltersFile_XX(t *testing.T) {
	rules.GlobalRemoteRuleFilters.Clear()
	defer rules.GlobalRemoteRuleFilters.Clear()

	n, err := LoadRemoteRuleFiltersFile(xxRemoteFiltersPath(t), "xx")
	require.NoError(t, err)
	require.Greater(t, n, 0)

	// TEST_REMOTE_RULE: surface "test" → drop remote match with that id on same span.
	sent := languagetool.AnalyzePlain("test")
	// Find token span for "test"
	from, to := 0, 4
	drop := rules.NewRuleMatch(rules.NewFakeRule("TEST_REMOTE_RULE"), sent, from, to, "d")
	keep := rules.NewRuleMatch(rules.NewFakeRule("OTHER"), sent, from, to, "k")
	out := rules.FilterRemoteRuleMatches("xx", sent, []*rules.RuleMatch{drop, keep})
	require.Len(t, out, 1)
	require.Equal(t, "OTHER", out[0].Rule.(*rules.FakeRule).GetID())

	// Position-aware: TEST_MARKER is foo + <marker>bar</marker> → span is bar only (Java).
	sent2 := languagetool.AnalyzePlain("foo bar")
	// bar at 4:7
	barOnly := rules.NewRuleMatch(rules.NewFakeRule("TEST_MARKER"), sent2, 4, 7, "m")
	out2 := rules.FilterRemoteRuleMatches("xx", sent2, []*rules.RuleMatch{barOnly})
	require.Empty(t, out2, "TEST_MARKER drops when remote span equals marker span")

	// Full pattern span must not drop (position equality required).
	whole := rules.NewRuleMatch(rules.NewFakeRule("TEST_MARKER"), sent2, 0, 7, "m")
	out3 := rules.FilterRemoteRuleMatches("xx", sent2, []*rules.RuleMatch{whole})
	require.Len(t, out3, 1)

	// ID regex full match: TEST_ID_REGEX[0-9]{1,2}
	sent3 := languagetool.AnalyzePlain("Lorem ipsum foo.")
	// find foo
	text := sent3.GetText()
	i := 0
	for j := 0; j+3 <= len(text); j++ {
		if text[j:j+3] == "foo" {
			i = j
			break
		}
	}
	ok1 := rules.NewRuleMatch(rules.NewFakeRule("TEST_ID_REGEX12"), sent3, i, i+3, "m")
	bad := rules.NewRuleMatch(rules.NewFakeRule("TEST_ID_REGEX123"), sent3, i, i+3, "m")
	out4 := rules.FilterRemoteRuleMatches("xx", sent3, []*rules.RuleMatch{ok1, bad})
	// ok1 dropped (id matches regex + span), bad kept (full-string regex fails for 123)
	require.Len(t, out4, 1)
	require.Equal(t, "TEST_ID_REGEX123", out4[0].Rule.(*rules.FakeRule).GetID())
}

func TestLoadRemoteRuleFiltersFile_Missing(t *testing.T) {
	n, err := LoadRemoteRuleFiltersFile("/no/such/remote-rule-filters.xml", "de")
	require.NoError(t, err)
	require.Equal(t, 0, n)
}
