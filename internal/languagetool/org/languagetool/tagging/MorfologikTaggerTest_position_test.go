package tagging_test

// Twin of MorfologikTaggerTest.testPositionWithIgnoredChars
// (Java uses Demo + JLanguageTool.getRawAnalyzedSentence — not MorfologikTagger itself).

import (
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/stretchr/testify/require"

	languagetool "github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// TestMorfologikTagger_PositionWithIgnoredChars twins
// MorfologikTaggerTest.testPositionWithIgnoredChars.
//
// Java Language base default ignoredCharactersRegex is [\u00AD]; Demo inherits it.
// NewJLanguageTool("xx") does not auto-wire that field yet, so the test sets the
// same default regex Java Demo would use.
//
// External package (tagging_test) avoids import cycle tagging ↔ languagetool.
func TestMorfologikTagger_PositionWithIgnoredChars(t *testing.T) {
	lt := languagetool.NewJLanguageTool("xx")
	// Java Language.ignoredCharactersRegex = [\u00AD]
	lt.IgnoredCharacters = languagetool.GermanIgnoredCharactersRegex

	text := "t\u00ADox te\u00ADstx"
	analyzedSent := lt.GetRawAnalyzedSentence(text)
	require.NotNil(t, analyzedSent)
	toks := analyzedSent.GetTokens()
	require.GreaterOrEqual(t, len(toks), 4)

	tok3 := toks[3]
	require.NotNil(t, tok3.GetToken())

	at0 := tok3.GetAnalyzedToken(0)
	require.NotNil(t, at0)
	require.NotNil(t, at0.GetPOSTag())
	require.Equal(t, "SENT_END", *at0.GetPOSTag())

	// Java String.indexOf / length are UTF-16 code units.
	require.Equal(t, utf16IndexOf(text, "te\u00ADst"), tok3.GetStartPos())
	require.Equal(t, utf16Len(text), tok3.GetEndPos())
}

func utf16IndexOf(s, sub string) int {
	b := strings.Index(s, sub)
	if b < 0 {
		return -1
	}
	return utf16Len(s[:b])
}

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}
