package commandline

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

type fakeChecker struct{ ms []*rules.RuleMatch }

func (f fakeChecker) Check(string) ([]*rules.RuleMatch, error) { return f.ms, nil }

func TestPrintMatchesAndCheckText(t *testing.T) {
	sent := languagetool.AnalyzePlain("I has a problem.")
	m := rules.NewRuleMatch(rules.NewFakeRule("DEMO_RULE"), sent, 2, 5, "grammar")
	m.SetSuggestedReplacements([]string{"have", "had"})

	var buf bytes.Buffer
	PrintMatches(&buf, []*rules.RuleMatch{m}, 0, "I has a problem.", 20, 0, false)
	out := buf.String()
	// PrintMatches ports CommandLineTools.printMatches (Rule ID / Message / Suggestion).
	require.Contains(t, out, "Rule ID: DEMO_RULE")
	require.Contains(t, out, "Message: grammar")
	require.Contains(t, out, "Suggestion: have; had")

	buf.Reset()
	n, err := CheckText(&buf, "I has a problem.", fakeChecker{ms: []*rules.RuleMatch{m}})
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Contains(t, buf.String(), "Time:")

	buf.Reset()
	n, err = CheckTextOpts(&buf, "x", fakeChecker{ms: nil}, CheckTextOptions{JSON: true})
	require.NoError(t, err)
	require.Equal(t, 0, n)
	require.Equal(t, "[]", strings.TrimSpace(buf.String()))
}

func TestDisplayTimeStatsAndProfile(t *testing.T) {
	var buf bytes.Buffer
	DisplayTimeStats(&buf, time.Now().Add(-100*time.Millisecond), 5)
	require.Contains(t, buf.String(), "sentences")

	buf.Reset()
	// Java profileRulesOnText: matchCount sums over 3 iterations × sentences.
	// 2 sentences × 1 match × 3 iterations = 6
	ProfileRulesOnText(&buf, []string{"a", "b"}, []string{"R1"}, func(ruleID, sentence string) int {
		return 1
	})
	out := buf.String()
	require.Contains(t, out, "R1")
	require.Contains(t, out, "6", "matchCount must sum across iterations like Java")
}

func TestLineColumnAt(t *testing.T) {
	line, col := lineColumnAt("ab\ncd", 4) // at 'd' (0-based index 4 is 'd')
	require.Equal(t, 2, line)
	require.Equal(t, 2, col)
}
