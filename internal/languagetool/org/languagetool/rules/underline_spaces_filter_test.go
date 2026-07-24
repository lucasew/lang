package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin helpers for UnderlineSpacesFilter.acceptRuleMatch (Java-visible outcomes only).

func underlineSpacesSentence(text string) *languagetool.AnalyzedSentence {
	tok := languagetool.NewAnalyzedToken(text, nil, nil)
	return languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(tok, 0),
	})
}

func underlineSpacesMatch(text string, from, to int) *RuleMatch {
	return NewRuleMatch(nil, underlineSpacesSentence(text), from, to, "msg")
}

func TestUnderlineSpacesFilter_AcceptRuleMatch_before(t *testing.T) {
	// "a b c" — "b" at UTF-16 [2,3); prev char is space → fromPos--
	f := NewUnderlineSpacesFilter()
	m := underlineSpacesMatch("a b c", 2, 3)
	out := f.AcceptRuleMatch(m, map[string]string{"underlineSpaces": "before"}, 0, nil, nil)
	require.Same(t, m, out)
	require.Equal(t, 1, out.GetFromPos())
	require.Equal(t, 3, out.GetToPos())
}

func TestUnderlineSpacesFilter_AcceptRuleMatch_after(t *testing.T) {
	// "a b c" length 5; toPos=3 → 3+1 < 5 and char at 3 is space → toPos++
	f := NewUnderlineSpacesFilter()
	m := underlineSpacesMatch("a b c", 2, 3)
	out := f.AcceptRuleMatch(m, map[string]string{"underlineSpaces": "after"}, 0, nil, nil)
	require.Same(t, m, out)
	require.Equal(t, 2, out.GetFromPos())
	require.Equal(t, 4, out.GetToPos())
}

func TestUnderlineSpacesFilter_AcceptRuleMatch_both(t *testing.T) {
	f := NewUnderlineSpacesFilter()
	m := underlineSpacesMatch("a b c", 2, 3)
	out := f.AcceptRuleMatch(m, map[string]string{"underlineSpaces": "both"}, 0, nil, nil)
	require.Same(t, m, out)
	require.Equal(t, 1, out.GetFromPos())
	require.Equal(t, 4, out.GetToPos())
}

func TestUnderlineSpacesFilter_AcceptRuleMatch_noWhitespaceAtEdges(t *testing.T) {
	f := NewUnderlineSpacesFilter()
	m := underlineSpacesMatch("abc", 1, 2)
	out := f.AcceptRuleMatch(m, map[string]string{"underlineSpaces": "both"}, 0, nil, nil)
	require.Same(t, m, out)
	require.Equal(t, 1, out.GetFromPos())
	require.Equal(t, 2, out.GetToPos())
}

func TestUnderlineSpacesFilter_AcceptRuleMatch_unknownModeUnchanged(t *testing.T) {
	// Java: only "before"/"after"/"both" expand; other values still return match
	f := NewUnderlineSpacesFilter()
	m := underlineSpacesMatch("a b c", 2, 3)
	out := f.AcceptRuleMatch(m, map[string]string{"underlineSpaces": "none"}, 0, nil, nil)
	require.Same(t, m, out)
	require.Equal(t, 2, out.GetFromPos())
	require.Equal(t, 3, out.GetToPos())
}

func TestUnderlineSpacesFilter_AcceptRuleMatch_missingKey(t *testing.T) {
	// Java RuleFilter.getRequired → IllegalArgumentException "Missing key 'underlineSpaces'"
	f := NewUnderlineSpacesFilter()
	m := underlineSpacesMatch("a b", 0, 1)
	require.PanicsWithValue(t, "Missing key 'underlineSpaces'", func() {
		f.AcceptRuleMatch(m, map[string]string{}, 0, nil, nil)
	})
	require.PanicsWithValue(t, "Missing key 'underlineSpaces'", func() {
		f.AcceptRuleMatch(m, nil, 0, nil, nil)
	})
}

func TestUnderlineSpacesFilter_AcceptRuleMatch_nullMatchOrSentenceNPE(t *testing.T) {
	// Java: match.getSentence().getText() → NPE
	f := NewUnderlineSpacesFilter()
	args := map[string]string{"underlineSpaces": "both"}
	require.PanicsWithValue(t, "NullPointerException", func() {
		f.AcceptRuleMatch(nil, args, 0, nil, nil)
	})
	m := NewRuleMatch(nil, nil, 0, 1, "msg")
	require.PanicsWithValue(t, "NullPointerException", func() {
		f.AcceptRuleMatch(m, args, 0, nil, nil)
	})
}

func TestUnderlineSpacesFilter_AcceptRuleMatch_bothOrderUsesUpdatedFromPos(t *testing.T) {
	// Java order: expand before first (setOffsetPosition), then after uses getFromPos()
	// which is already reduced. Final span must include both sides.
	f := NewUnderlineSpacesFilter()
	m := underlineSpacesMatch("a b c", 2, 3)
	out := f.AcceptRuleMatch(m, map[string]string{"underlineSpaces": "both"}, 0, nil, nil)
	require.Equal(t, 1, out.GetFromPos())
	require.Equal(t, 4, out.GetToPos())

	// after-only on already-before-expanded positions still works
	m2 := underlineSpacesMatch("a b c", 1, 3) // already includes leading space
	out2 := f.AcceptRuleMatch(m2, map[string]string{"underlineSpaces": "after"}, 0, nil, nil)
	require.Equal(t, 1, out2.GetFromPos())
	require.Equal(t, 4, out2.GetToPos())
}

func TestUnderlineSpacesFilter_AcceptRuleMatch_javaAfterLengthQuirk(t *testing.T) {
	// Java: toPos+1 < sentence.length() — trailing single whitespace is NOT expanded.
	// "a b " length 4; match "b" [2,3); 3+1 < 4 is false.
	f := NewUnderlineSpacesFilter()
	m := underlineSpacesMatch("a b ", 2, 3)
	out := f.AcceptRuleMatch(m, map[string]string{"underlineSpaces": "after"}, 0, nil, nil)
	require.Equal(t, 2, out.GetFromPos())
	require.Equal(t, 3, out.GetToPos())

	// With one more char after the space: "a b  " length 5; 3+1 < 5 → expand once.
	m2 := underlineSpacesMatch("a b  ", 2, 3)
	out2 := f.AcceptRuleMatch(m2, map[string]string{"underlineSpaces": "after"}, 0, nil, nil)
	require.Equal(t, 2, out2.GetFromPos())
	require.Equal(t, 4, out2.GetToPos())
}

func TestUnderlineSpacesFilter_AcceptRuleMatch_beforeAtStart(t *testing.T) {
	// fromPos-1 < 0 → no expand
	f := NewUnderlineSpacesFilter()
	m := underlineSpacesMatch(" ab", 0, 1)
	out := f.AcceptRuleMatch(m, map[string]string{"underlineSpaces": "before"}, 0, nil, nil)
	require.Equal(t, 0, out.GetFromPos())
	require.Equal(t, 1, out.GetToPos())
}

func TestUnderlineSpacesFilter_AcceptRuleMatch_utf16BMPIndexing(t *testing.T) {
	// NBSP (U+00A0) is one UTF-16 unit and IsWhitespace-true.
	// "x\u00A0y" — match "y" at [2,3); before expands over NBSP.
	f := NewUnderlineSpacesFilter()
	m := underlineSpacesMatch("x\u00A0y z", 2, 3)
	out := f.AcceptRuleMatch(m, map[string]string{"underlineSpaces": "both"}, 0, nil, nil)
	require.Equal(t, 1, out.GetFromPos())
	require.Equal(t, 4, out.GetToPos()) // after space before 'z'
}
