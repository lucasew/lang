package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXMLRuleHandlerFinishRule(t *testing.T) {
	h := NewXMLRuleHandler("en")
	h.ID = "R1"
	h.Name = "Rule one"
	h.PatternTokens = []*PatternToken{Token("hello")}
	h.Message = "bad"
	r := h.FinishRule()
	require.NotNil(t, r)
	require.Equal(t, "R1", r.ID)
	require.Len(t, h.GetRules(), 1)
	require.True(t, AttrYes("yes"))
	require.False(t, AttrYes("no"))
}
