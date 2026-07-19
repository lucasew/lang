package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrefixLists_FromJava(t *testing.T) {
	require.GreaterOrEqual(t, len(prefixesSeparableVerbsJava), 150)
	require.GreaterOrEqual(t, len(prefixesVerbsJava), 150)
	// longest-first: first entry longer than last short ones
	require.Greater(t, len(prefixesSeparableVerbsLongestList[0]), len(prefixesSeparableVerbsLongestList[len(prefixesSeparableVerbsLongestList)-1]))
	require.True(t, isExactSeparablePrefix("ein"))
	require.True(t, isExactSeparablePrefix("zurück"))
	require.False(t, isExactSeparablePrefix("xyz"))
	// non-separable present in verbs list
	require.Contains(t, prefixesVerbsJava, "ver")
	require.Contains(t, prefixesVerbsJava, "be")
}

func TestDomainLikeSequence(t *testing.T) {
	require.True(t, isDomainLikeSequence([]string{"example", ".", "com"}, 0))
	require.True(t, isDomainLikeSequence([]string{"foo", ".", "DE"}, 0))
	require.False(t, isDomainLikeSequence([]string{"example", "com"}, 0))
	require.False(t, isDomainLikeSequence([]string{"example", ".", "xyz"}, 0))
}
