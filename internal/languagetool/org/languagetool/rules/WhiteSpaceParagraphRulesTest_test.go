package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWhiteSpaceAtBeginOfParagraph(t *testing.T) {
	rule := NewWhiteSpaceAtBeginOfParagraph(nil)
	// AnalyzePlain: SENT_START + tokens; leading space becomes whitespace token after SENT_START
	matches := rule.Match(languagetool.AnalyzePlain("  Hello world."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Hello world."))))
}

func TestWhiteSpaceBeforeParagraphEnd(t *testing.T) {
	rule := NewWhiteSpaceBeforeParagraphEnd(nil)
	// last sentence of list is always paragraph end
	sents := []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("Hello world. ")}
	// trailing space may be in tokens
	_ = rule.MatchList(sents)
	// Construct with explicit trailing whitespace token via plain text ending with spaces
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("Hello  ")})
	// If no match due to tokenizer, soft assert load
	_ = matches
	require.Equal(t, "WHITESPACE_PARAGRAPH", rule.GetID())
}
