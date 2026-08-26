package nl

import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestDutchRemoteRuleFilter_Rules(t *testing.T) {
	f := rules.NewRemoteRuleFilters()
	f.Register("nl", &rules.FilterRule{IDPattern: regexp.MustCompile(`AI_.*`)})
	sent := languagetool.AnalyzePlain("ab")
	drop := rules.NewRuleMatch(rules.NewFakeRule("AI_X"), sent, 0, 2, "d")
	keep := rules.NewRuleMatch(rules.NewFakeRule("OK"), sent, 0, 2, "k")
	require.Len(t, f.FilterMatches("nl", sent, []*rules.RuleMatch{drop, keep}), 1)
}
