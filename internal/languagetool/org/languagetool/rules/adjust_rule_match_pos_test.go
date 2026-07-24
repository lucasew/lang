package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
	"github.com/stretchr/testify/require"
)

func TestRuleMatch_LineColumnPatternDefaults(t *testing.T) {
	rm := NewRuleMatch(NewFakeRule("R"), nil, 2, 5, "msg")
	require.Equal(t, -1, rm.GetLine())
	require.Equal(t, -1, rm.GetColumn())
	require.Equal(t, 2, rm.GetPatternFromPos())
	require.Equal(t, 5, rm.GetPatternToPos())
	// unset sentence pos falls back to document offset (Java getFromPosSentence)
	require.Equal(t, 2, rm.GetFromPosSentence())
	require.Equal(t, 5, rm.GetToPosSentence())
	rm.SetLine(3)
	rm.SetEndLine(4)
	rm.SetColumn(1)
	rm.SetEndColumn(8)
	require.Equal(t, 3, rm.GetLine())
	require.Equal(t, 4, rm.GetEndLine())
	require.Equal(t, 1, rm.GetColumn())
	require.Equal(t, 8, rm.GetEndColumn())
}

func TestAdjustRuleMatchPos_Basic(t *testing.T) {
	// sentence "Hello world" — match "world" at 6..11
	rm := NewRuleMatch(NewFakeRule("R"), nil, 6, 11, "msg")
	// prior sentences: charCount 10, columnCount 0, lineCount 1
	adj := AdjustRuleMatchPos(rm, 10, 0, 1, "Hello world", nil)
	require.Equal(t, 16, adj.GetFromPos())
	require.Equal(t, 21, adj.GetToPos())
	require.Equal(t, 6, adj.GetFromPosSentence())
	require.Equal(t, 11, adj.GetToPosSentence())
	require.Equal(t, 16, adj.GetPatternFromPos())
	require.Equal(t, 21, adj.GetPatternToPos())
	// no newline in prefix → column = 6 + 0 = 6, endColumn = 11
	require.Equal(t, 1, adj.GetLine())
	require.Equal(t, 1, adj.GetEndLine())
	require.Equal(t, 6, adj.GetColumn())
	require.Equal(t, 11, adj.GetEndColumn())
}

func TestAdjustRuleMatchPos_WithNewline(t *testing.T) {
	// "ab\ncde" match "cde" at 3..6
	rm := NewRuleMatch(NewFakeRule("R"), nil, 3, 6, "msg")
	adj := AdjustRuleMatchPos(rm, 0, 5, 2, "ab\ncde", nil)
	// prefix to error "ab\n" → last nl at 2, column = 3-2 = 1
	require.Equal(t, 1, adj.GetColumn())
	// prefix to end "ab\ncde" last nl at 2, endColumn = 6-2 = 4
	require.Equal(t, 4, adj.GetEndColumn())
	require.Equal(t, 3, adj.GetLine())    // 2 + 1 linebreak
	require.Equal(t, 3, adj.GetEndLine()) // 2 + 1
}

func TestAdjustRuleMatchPos_AnnotatedText(t *testing.T) {
	// plain "foo" with markup before: mapping like AnnotatedTextTest
	b := markup.NewAnnotatedTextBuilder()
	b.AddMarkup("<b>")
	b.AddText("foo")
	b.AddMarkup("</b>")
	at := b.Build()
	// match whole "foo" 0..3 in plain
	rm := NewRuleMatch(NewFakeRule("R"), nil, 0, 3, "msg")
	adj := AdjustRuleMatchPos(rm, 0, 0, 0, "foo", at)
	// original positions should map past markup
	require.Greater(t, adj.GetFromPos(), 0) // after <b>
	require.Greater(t, adj.GetToPos(), adj.GetFromPos())
}

func TestCloneRuleMatch(t *testing.T) {
	rm := NewRuleMatch(NewFakeRule("R"), languagetool.AnalyzePlain("hi"), 0, 2, "m")
	rm.SetSuggestedReplacements([]string{"a"})
	rm.SetLine(1)
	c := CloneRuleMatch(rm)
	require.Equal(t, rm.GetFromPos(), c.GetFromPos())
	require.Equal(t, 1, c.GetLine())
	require.Equal(t, []string{"a"}, c.GetSuggestedReplacements())
	c.SetLine(9)
	require.Equal(t, 1, rm.GetLine())
}
