package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestPlainRuleSet(t *testing.T) {
	r1 := rules.NewFakeRule("A")
	r2 := rules.NewFakeRule("B")
	set := PlainRuleSet([]RuleIDGetter{r1, r2})
	require.Len(t, set.AllRules(), 2)
	require.Len(t, set.RulesForSentence(nil), 2)
	_, ok := set.AllRuleIDs()["A"]
	require.True(t, ok)
	_, ok = set.AllRuleIDs()["B"]
	require.True(t, ok)
}
