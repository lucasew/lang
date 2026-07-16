package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEmptyLineRule(t *testing.T) {
	rule := NewEmptyLineRule(nil)
	require.Equal(t, "EMPTY_LINE", rule.GetID())
	// Single-line-break languages: empty line is \n\n at paragraph end.
	rule.SingleLineBreaksMarksPara = true
	require.True(t, rule.isSecondParagraphEndMark("Hello.\n\n"))
	require.False(t, rule.isSecondParagraphEndMark("Hello.\n"))
	// Default mode: four newlines
	rule.SingleLineBreaksMarksPara = false
	require.True(t, rule.isSecondParagraphEndMark("Hello.\n\n\n\n"))
	require.False(t, rule.isSecondParagraphEndMark("Hello.\n\n"))
}

func TestEmptyLineRule_MatchList(t *testing.T) {
	rule := NewEmptyLineRule(nil)
	rule.SingleLineBreaksMarksPara = true
	// First sentence is paragraph end if next starts with \n; also needs text ending \n\n
	// AnalyzeTextDemo may put newlines into sentence text.
	sents := languagetool.AnalyzeTextDemo("Hello world.\n\n\nNext para.")
	_ = rule.MatchList(sents) // no panic
}
