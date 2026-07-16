package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.ConfusionPairTest.

func TestConfusionPair_Get(t *testing.T) {
	cs := NewConfusionPairTokens("one", "two", 1, true)
	require.Len(t, cs.GetTerms(), 2)
	require.True(t, strings.Contains(termsString(cs.GetTerms()), "one"))
	require.True(t, strings.Contains(termsString(cs.GetTerms()), "two"))
	require.Len(t, cs.GetUppercaseFirstCharTerms(), 2)
	require.True(t, strings.Contains(termsString(cs.GetUppercaseFirstCharTerms()), "One"))
	require.True(t, strings.Contains(termsString(cs.GetUppercaseFirstCharTerms()), "Two"))
}

func TestConfusionPair_Equals(t *testing.T) {
	a := NewConfusionPairTokens("one", "two", 1, true)
	a2 := NewConfusionPairTokens("one", "two", 1, true)
	b := NewConfusionPairTokens("two", "one", 1, true)
	c := NewConfusionPairTokens("Two", "one", 1, true)
	d := NewConfusionPairTokens("Two", "one", 2, true)
	require.True(t, a.Equals(a2))
	require.False(t, a.Equals(b))
	require.False(t, a.Equals(c))
	require.False(t, b.Equals(c))
	require.False(t, c.Equals(d))
}

func termsString(terms []*ConfusionString) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, t := range terms {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(t.String())
	}
	b.WriteByte(']')
	return b.String()
}
