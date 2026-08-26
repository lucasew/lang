package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatternRuleId(t *testing.T) {
	p := NewPatternRuleId("FOO")
	require.Equal(t, "FOO", p.GetID())
	require.Nil(t, p.GetSubID())
	require.Equal(t, "FOO", p.String())

	p2 := NewPatternRuleIdWithSub("FOO", "2")
	require.Equal(t, "FOO", p2.GetID())
	require.Equal(t, "2", *p2.GetSubID())
	require.Equal(t, "FOO[2]", p2.String())

	require.Panics(t, func() { NewPatternRuleId("") })
	require.Panics(t, func() { NewPatternRuleIdWithSub("x", "") })
}
