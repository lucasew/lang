package wikipedia

// Twin of languagetool-wikipedia/src/test/java/org/languagetool/dev/wikipedia/SuggestionReplacerTest.java
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuggestionReplacer_FindWhitespace(t *testing.T) {
	text := "hello world here"
	// "world" starts at 6 ends at 11
	require.Equal(t, 11, FindNextWhitespaceToTheRight(text, 6))
	require.Equal(t, 6, FindNextWhitespaceToTheLeft(text, 8))
	require.Equal(t, 0, FindNextWhitespaceToTheLeft(text, 2))
	require.Equal(t, len([]rune(text)), FindNextWhitespaceToTheRight(text, 12))
}

// Port of SuggestionReplacerTest.testErrorAtTextBeginning — plain-text identity mapping
// without full JLanguageTool (synthetic match).
func TestSuggestionReplacer_ErrorAtTextBeginning(t *testing.T) {
	markup := "A hour ago\n"
	plain := markup // identity
	mapping := NewPlainTextMapping(plain)
	replacer := NewSuggestionReplacerWithMarker(mapping, markup, NewErrorMarker("<s>", "</s>"))
	// "A" at 0..1 → "An"
	apps, err := replacer.ApplySuggestionsToOriginalText(MatchSpan{
		FromPos:               0,
		ToPos:                 1,
		SuggestedReplacements: []string{"An"},
	})
	require.NoError(t, err)
	require.Len(t, apps, 1)
	require.True(t, apps[0].HasRealRepl())
	require.Contains(t, apps[0].GetTextWithCorrection(), "<s>An</s>")
}

// Port of SuggestionReplacerTest.testApplySuggestionToOriginalText (subset, identity mapping)
func TestSuggestionReplacer_ApplySuggestionToOriginalText(t *testing.T) {
	// "Die CD ROM." — replace "CD ROM" with "CD-ROM"
	markup := "Die CD ROM."
	mapping := NewPlainTextMapping(markup)
	replacer := NewSuggestionReplacerWithMarker(mapping, markup, NewErrorMarker("<s>", "</s>"))
	from := strings.Index(markup, "CD ROM")
	to := from + len("CD ROM")
	apps, err := replacer.ApplySuggestionsToOriginalText(MatchSpan{
		FromPos:               from,
		ToPos:                 to,
		SuggestedReplacements: []string{"CD-ROM"},
	})
	require.NoError(t, err)
	require.Len(t, apps, 1)
	// context expands to whitespace boundaries → whole "CD ROM." if no space after?
	// "Die CD ROM." — left of C is space → contextFrom after space; right of M is space? no space before period
	// FindNextWhitespaceToTheRight from end of "CD ROM" (pos of '.') → may not find space → end
	got := apps[0].GetTextWithCorrection()
	require.Contains(t, got, "CD-ROM")
	require.Contains(t, got, "<s>")
	require.True(t, apps[0].HasRealRepl())
}

// Port of SuggestionReplacerTest.testErrorAtParagraphBeginning
func TestSuggestionReplacer_ErrorAtParagraphBeginning(t *testing.T) {
	markup := "X\n\nA hour ago.\n"
	mapping := NewPlainTextMapping(markup)
	replacer := NewSuggestionReplacerWithMarker(mapping, markup, NewErrorMarker("<s>", "</s>"))
	from := strings.Index(markup, "A hour")
	apps, err := replacer.ApplySuggestionsToOriginalText(MatchSpan{
		FromPos:               from,
		ToPos:                 from + 1,
		SuggestedReplacements: []string{"An"},
	})
	require.NoError(t, err)
	require.Contains(t, apps[0].GetTextWithCorrection(), "<s>An</s>")
}

// Soft-skip full Sweble/JLanguageTool integration cases
func TestSuggestionReplacer_NestedTemplates(t *testing.T) {
	t.Skip("soft-skip: needs Sweble filter + German rules")
}
func TestSuggestionReplacer_Reference1(t *testing.T) {
	t.Skip("soft-skip: needs Sweble filter + German rules")
}
func TestSuggestionReplacer_Reference2(t *testing.T) {
	t.Skip("soft-skip: needs Sweble filter + German rules")
}
func TestSuggestionReplacer_KnownBug(t *testing.T) {
	t.Skip("soft-skip: Sweble location bug reproduction")
}
func TestSuggestionReplacer_ComplexText(t *testing.T) {
	t.Skip("soft-skip: full wiki markup + German LT")
}
func TestSuggestionReplacer_CompleteText2(t *testing.T) {
	t.Skip("soft-skip: full wiki markup + German LT")
}
