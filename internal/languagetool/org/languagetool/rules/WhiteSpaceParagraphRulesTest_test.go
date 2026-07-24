package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// WhiteSpaceAtBeginOfParagraph: see WhiteSpaceAtBeginOfParagraphTest_test.go
// (demoLT/EN Analyze behavior matrix; no invent AnalyzePlain as sole proof).

// WhiteSpaceBeforeParagraphEnd is a separate leaf; keep package tests compiling only.
func TestWhiteSpaceBeforeParagraphEnd(t *testing.T) {
	rule := NewWhiteSpaceBeforeParagraphEnd(nil)
	// Soft load path for package compile; full twin is a different checklist leaf.
	sents := []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("Hello  ")}
	_ = rule.MatchList(sents)
	require.Equal(t, "WHITESPACE_PARAGRAPH", rule.GetID())
}
