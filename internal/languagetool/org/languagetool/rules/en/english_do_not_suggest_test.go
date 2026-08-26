package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnglishDoNotSuggestWords_FullList(t *testing.T) {
	// subset of Java lcDoNotSuggestWords
	require.True(t, IsDoNotSuggest("bullshit"))
	require.True(t, IsDoNotSuggest("Bullshit"))
	require.True(t, IsDoNotSuggest("niggardly"))
	require.True(t, IsDoNotSuggest("double check"))
	require.True(t, IsDoNotSuggest("in house"))
	require.False(t, IsDoNotSuggest("hello"))
	// map should be larger than old subset
	require.Greater(t, len(EnglishDoNotSuggestWords), 50)
}

func TestEnglishCheckCompoundEnabled(t *testing.T) {
	r := NewAbstractEnglishSpellerRule("MORFOLOGIK_RULE_EN_US", "en-US", nil)
	require.True(t, r.CheckCompound)
}
