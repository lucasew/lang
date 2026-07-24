package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Java AdvancedWordRepeatRule: token.length() < 2 uses UTF-16 units.
// A single non-BMP code point has length 2 in Java and is still a "word".
func TestAdvancedWordRepeatRule_UTF16TokenLengthGate(t *testing.T) {
	r := &AdvancedWordRepeatRule{
		ID:             "ADV_REPEAT",
		Message:        "repeat",
		ExcludedWords:  map[string]bool{},
	}
	// emoji is 1 rune / 2 UTF-16 units — Java length is 2 → isWord stays true
	emoji := "😀"
	require.Equal(t, 2, utf16LenAdv(emoji))
	// two identical emoji tokens → surface repeat when untagged (hasLemma false)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(emoji, nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(emoji, nil, nil), 2),
	})
	ms := r.Match(sent)
	require.Len(t, ms, 1, "Java treats length-2 surrogate pair as word; repeated emoji matches")
	require.Equal(t, 2, ms[0].FromPos)
	require.Equal(t, 4, ms[0].ToPos) // start 2 + UTF-16 length 2

	// ASCII single letter still excluded (length 1)
	r2 := &AdvancedWordRepeatRule{ID: "ADV_REPEAT", Message: "repeat", ExcludedWords: map[string]bool{}}
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("a", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("a", nil, nil), 1),
	})
	require.Empty(t, r2.Match(sent2), "single-letter ASCII not a word (length < 2)")
}
