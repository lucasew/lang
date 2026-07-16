package rules

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.PunctuationMarkAtParagraphEndTest.

func paraEndRule() *PunctuationMarkAtParagraphEnd {
	return NewPunctuationMarkAtParagraphEnd(map[string]string{
		"punctuation_mark_paragraph_end_msg":  "Add a punctuation mark at paragraph end",
		"punctuation_mark_paragraph_end_desc": "Paragraph needs final punctuation",
	})
}

// analyzeForParaEnd keeps sentence-local positions (TextLevelRule adds corrected length).
// Paragraph breaks: \n\n. Sentence splits on .!? like SplitAndAnalyze, but without
// pre-shifting positions.
func analyzeForParaEnd(input string) []*languagetool.AnalyzedSentence {
	// Use AnalyzeTextDemo for \n\n paragraphs; for single-line multi-sentence use SplitAndAnalyze
	// but re-localize positions (zero-based per sentence).
	var raw []*languagetool.AnalyzedSentence
	if strings.Contains(input, "\n\n") {
		raw = languagetool.AnalyzeTextDemo(input)
	} else if strings.Contains(input, ". ") || strings.Contains(input, ".\n") ||
		strings.Contains(input, "! ") || strings.Contains(input, "? ") ||
		// "2." list markers / trailing sentence fragments
		strings.Contains(input, ".") {
		raw = languagetool.SplitAndAnalyze(input)
	} else {
		raw = []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(input)}
	}
	// Un-shift to sentence-local positions so MatchList's pos accumulation is correct.
	for _, s := range raw {
		toks := s.GetTokens()
		if len(toks) == 0 {
			continue
		}
		// Find min start among non-SENT_START tokens; SENT_START is 0.
		base := 0
		for _, t := range toks {
			if t.IsSentenceStart() {
				continue
			}
			base = t.GetStartPos()
			break
		}
		if base == 0 {
			continue
		}
		for _, t := range toks {
			t.SetStartPos(t.GetStartPos() - base)
		}
	}
	return raw
}

func assertParaEndMatches(t *testing.T, expected int, input string) {
	t.Helper()
	rule := paraEndRule()
	got := len(rule.MatchList(analyzeForParaEnd(input)))
	require.Equal(t, expected, got, "input=%q", input)
}

func TestPunctuationMarkAtParagraphEnd_Rule(t *testing.T) {
	assertParaEndMatches(t, 0, "A paragraph.\n2. Some headline\n\n(a) A new sentence.")
	assertParaEndMatches(t, 0, "A paragraph.\n\n2. Some headline\n\n(a) A new sentence.")
	assertParaEndMatches(t, 0, "A paragraph.\n2.2.1 Some headline\n\n(a) A new sentence.")
	assertParaEndMatches(t, 0, "A paragraph.\n\n2.2.1 Some headline\n\n(a) A new sentence.")
	assertParaEndMatches(t, 0, "2. This is an item in a list")
	assertParaEndMatches(t, 0, "2.2.2. This is an item in a list")
	assertParaEndMatches(t, 0, "a) This is an item in a list")
	assertParaEndMatches(t, 0, "a.) This is an item in a list")
	assertParaEndMatches(t, 0, "\u2713 This is an item in a list")
	assertParaEndMatches(t, 0, "* This is an item in a list")
	assertParaEndMatches(t, 0, "This is a test sentence.")
	assertParaEndMatches(t, 0, "This is a test headline")
	assertParaEndMatches(t, 0, "This is a test sentence. This is a link: http://example.com")
	assertParaEndMatches(t, 1, "This is a test sentence. It can be found at http://example.com/foobar")
	assertParaEndMatches(t, 1, "This is a test sentence. And this is a second test sentence")
	assertParaEndMatches(t, 1, "\"This is a test sentence. And this is a second test sentence")
	assertParaEndMatches(t, 0, "This is a test sentence. And this is a second test sentence.")
	assertParaEndMatches(t, 0, "B. v. – Beschluss vom")
	assertParaEndMatches(t, 1,
		"This is a test sentence.\nAnd this is a second test sentence. Here is a dot missing")
	assertParaEndMatches(t, 0,
		"This is a test sentence.\nAnd this is a second test sentence. Here is a dot missing.")
	assertParaEndMatches(t, 0,
		"This is a sentence. Another one: https://languagetool.org/foo\n\nAnother sentence\n")
}
