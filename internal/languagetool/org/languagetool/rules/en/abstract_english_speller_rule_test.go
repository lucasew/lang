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

// Java AbstractEnglishSpellerRule: sentenc → sentence example pair.
func TestAbstractEnglishSpellerRule_ExamplePair(t *testing.T) {
	r := NewAbstractEnglishSpellerRule("MORFOLOGIK_RULE_EN_US", "en-US", nil)
	inc := r.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "This <marker>sentenc</marker> contains a spelling mistake.", inc[0].GetExample())
	require.Equal(t, []string{"sentence"}, inc[0].GetCorrections())
}
