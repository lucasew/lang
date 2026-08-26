package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuleMatchListener(t *testing.T) {
	var got []*RuleMatch
	l := RuleMatchListenerFunc(func(m *RuleMatch) { got = append(got, m) })
	m := NewRuleMatch(nil, nil, 0, 1, "x")
	NotifyListeners(m, l)
	require.Len(t, got, 1)
	require.Equal(t, "x", got[0].Message)
}
