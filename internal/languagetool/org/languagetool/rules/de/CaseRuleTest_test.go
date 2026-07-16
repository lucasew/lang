package de

// Twin of CaseRuleTest (surface capitalization heuristics).
import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCaseRule_Rule(t *testing.T) {
	rule := NewCaseRule(nil)
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	// goods
	require.Equal(t, 0, matchN("Ein einfacher Satz zum Testen."))
	require.Equal(t, 0, matchN("Heute spricht Frau Stieg."))
	// bads
	require.Equal(t, 1, matchN("Und das Neue Haus."))
	require.Equal(t, 1, matchN("Das sind die Die Lehrer."))
	require.Equal(t, 1, matchN("Ich habe Heute keine Zeit."))
	require.GreaterOrEqual(t, matchN("Ich wünsche dir Alles Liebe."), 1)
}

func TestCaseRuleCompareLists(t *testing.T) {
	s := languagetool.AnalyzePlain("Hier ein Test")
	toks := s.GetTokensWithoutWhitespace()
	// tokens: "", Hier, ein, Test, .
	require.True(t, CaseRuleCompareLists(toks, 1, 2,
		[]*regexp.Regexp{regexp.MustCompile("Hier"), regexp.MustCompile("ein")}))
	require.False(t, CaseRuleCompareLists(toks, 1, 2,
		[]*regexp.Regexp{regexp.MustCompile("Hier"), regexp.MustCompile("Test")}))
}
