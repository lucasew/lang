package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAbstractEnglishSpellerRule(t *testing.T) {
	r := NewAbstractEnglishSpellerRule("MORFOLOGIK_RULE_EN_US", "en-US", nil)
	files := r.GetAdditionalSpellingFileNames()
	require.Contains(t, files, EnglishMultiwordsFile)
	require.Contains(t, files, EnglishGlobalSpellingFile)
	require.True(t, IsDoNotSuggest("bullshit"))
	require.False(t, IsDoNotSuggest("hello"))
	require.Equal(t, []string{"hello"}, FilterEnglishSuggestions([]string{"hello", "bullshit"}))
}
