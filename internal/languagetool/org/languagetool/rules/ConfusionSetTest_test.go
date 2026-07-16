package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.ConfusionSetTest.

func TestConfusionSet_Get(t *testing.T) {
	cs := NewConfusionSet(1, "one", "two")
	require.Len(t, cs.GetSet(), 2)
	require.True(t, strings.Contains(cs.String(), "one"))
	require.True(t, strings.Contains(cs.String(), "two"))
	require.Len(t, cs.GetUppercaseFirstCharSet(), 2)
	up := termsString(cs.GetUppercaseFirstCharSet())
	require.True(t, strings.Contains(up, "One"))
	require.True(t, strings.Contains(up, "Two"))
}

func TestConfusionSet_Equals(t *testing.T) {
	a := NewConfusionSet(1, "one", "two")
	b := NewConfusionSet(1, "two", "one")
	c := NewConfusionSet(1, "Two", "one")
	d := NewConfusionSet(2, "Two", "one")
	require.True(t, a.Equals(b))
	require.False(t, a.Equals(c))
	require.False(t, b.Equals(c))
	require.False(t, c.Equals(d))
}
