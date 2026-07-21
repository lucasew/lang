package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
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

// matchSurfaceEN uses UTF-16 indices (Java String.substring), not runes/bytes.
func TestMatchSurfaceEN_UTF16(t *testing.T) {
	// "café X" — é is one UTF-16 unit; "X" at UTF-16 index 5
	text := "café X"
	sent := languagetool.AnalyzePlain(text)
	// positions 5,6 for "X"
	m := rules.NewRuleMatch(rules.NewFakeRule("T"), sent, 5, 6, "x")
	require.Equal(t, "X", matchSurfaceEN(m, sent))
	// "café" is UTF-16 [0:4]
	m2 := rules.NewRuleMatch(rules.NewFakeRule("T"), sent, 0, 4, "c")
	require.Equal(t, "café", matchSurfaceEN(m2, sent))
}
