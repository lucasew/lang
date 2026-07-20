package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuleMatchListener(t *testing.T) {
	var got any
	var l RuleMatchListener = func(match any) { got = match }
	l.MatchFound("m1")
	require.Equal(t, "m1", got)
	var nilL RuleMatchListener
	require.NotPanics(t, func() { nilL.MatchFound("x") })
}
