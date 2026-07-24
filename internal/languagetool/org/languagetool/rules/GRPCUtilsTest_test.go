package rules

// Twin of GRPCUtilsTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of GRPCUtilsTest.testLevels
func TestGRPCUtils_Levels(t *testing.T) {
	for _, l := range []CheckLevel{CheckLevelDefault, CheckLevelPicky} {
		g := LevelToGRPC(l)
		require.NotEmpty(t, g)
		require.Equal(t, l, LevelFromGRPC(g))
	}
	require.Equal(t, CheckLevelDefault, LevelFromGRPC("UNRECOGNIZED"))
}

// Port of GRPCUtilsTest.testURLFromRule
func TestGRPCUtils_URLFromRule(t *testing.T) {
	s := languagetool.AnalyzePlain("This is a test")
	rule := NewFakeRule("FAKE")
	rule.SetURL("http://example.com/")
	m := NewRuleMatch(rule, s, 0, 1, "test")
	g := MatchToGRPC(m)
	require.Equal(t, "http://example.com/", g.URL)
}

// Port of GRPCUtilsTest.testURLFromRuleMatch
func TestGRPCUtils_URLFromRuleMatch(t *testing.T) {
	s := languagetool.AnalyzePlain("This is a test")
	rule := NewFakeRule("FAKE")
	rule.SetURL("http://example.com/wrong")
	m := NewRuleMatch(rule, s, 0, 1, "test")
	m.SetURL("http://example.com/")
	g := MatchToGRPC(m)
	require.Equal(t, "http://example.com/", g.URL)
}
