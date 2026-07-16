package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIPOSTag(t *testing.T) {
	require.True(t, IPOSNoun.Match("noun:m:v_naz"))
	require.False(t, IPOSVerb.Match("noun:m"))
	require.True(t, IsNum("numr:p:v_naz"))
	require.True(t, IsNum("number"))
	require.True(t, POSContains("noun:m:v_rod", "v_rod"))
	require.True(t, POSStartsWithAny("verb:imperf", IPOSVerb, IPOSNoun))
	require.False(t, POSStartsWithAny("adv", IPOSVerb))
}

func TestLetterEndingNumeric(t *testing.T) {
	cases := CasesForNumericEnding("1", "й")
	require.Contains(t, cases, ":m:v_naz")
	require.Nil(t, CasesForNumericEnding("1", "zzz"))
}
