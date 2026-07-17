package pl

// Twin of MatchTest (Polish) — case/regex Match surface without full synthesizer.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Port of MatchTest.testSpeller (soft: Match construction + case conversion)
func TestMatch_Speller(t *testing.T) {
	m := patterns.NewMatch("subst:.*", "subst:sg:nom:m", true, "", "", patterns.CaseNone, false, true, patterns.IncludeNone)
	require.True(t, m.ChecksSpelling())
	require.True(t, m.IsPostagRegexp())
	require.Equal(t, "subst:.*", m.GetPosTag())
	// text match with regex replace
	tm := patterns.NewMatch("", "", false, "a(.)", "b$1", patterns.CaseStartUpper, false, false, patterns.IncludeNone)
	require.True(t, tm.ConvertsCase())
	require.Equal(t, "Hello", patterns.ConvertCase(patterns.CaseStartUpper, "hello", "X"))
}
