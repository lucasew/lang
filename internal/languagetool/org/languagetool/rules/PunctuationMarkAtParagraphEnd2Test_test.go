package rules

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.PunctuationMarkAtParagraphEnd2Test.

func paraEnd2Rule() *PunctuationMarkAtParagraphEnd2 {
	return NewPunctuationMarkAtParagraphEnd2(map[string]string{
		"punctuation_mark_paragraph_end_msg":  "Add a punctuation mark at paragraph end",
		"punctuation_mark_paragraph_end_desc": "Paragraph needs final punctuation",
	})
}

// analyzeForParaEnd2: sentence-local positions; \n\n paragraphs; .!? splits
// without pre-shifting (MatchList accumulates corrected length).
func analyzeForParaEnd2(input string) []*languagetool.AnalyzedSentence {
	var raw []*languagetool.AnalyzedSentence
	if strings.Contains(input, "\n\n") {
		raw = languagetool.AnalyzeTextDemo(input)
	} else if strings.Contains(input, ". ") || strings.Contains(input, ".\n") ||
		strings.Contains(input, "! ") || strings.Contains(input, "? ") ||
		// list markers / multi-sentence
		(strings.Contains(input, ".") && (strings.Contains(input, ". ") || isLikelyMultiSentence(input))) {
		raw = languagetool.SplitAndAnalyze(input)
	} else {
		raw = []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(input)}
	}
	// localize positions
	for _, s := range raw {
		toks := s.GetTokens()
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
	// trailing single \n on last sentence: keep as part of text for linebreak tokens
	return raw
}

func isLikelyMultiSentence(input string) bool {
	// "2." list or "text. Text"
	for i := 0; i < len(input)-1; i++ {
		if input[i] == '.' {
			return true
		}
	}
	return false
}

func assertParaEnd2(t *testing.T, expected int, input string) {
	t.Helper()
	got := len(paraEnd2Rule().MatchList(analyzeForParaEnd2(input)))
	require.Equal(t, expected, got, "input=%q", input)
}

func TestPunctuationMarkAtParagraphEnd2_Test(t *testing.T) {
	assertParaEnd2(t, 0, "2. This is an item in a list")
	assertParaEnd2(t, 0, "2.2.2. This is an item in a list")
	assertParaEnd2(t, 0, "This is a test.")
	assertParaEnd2(t, 0, "This is a test") // too short
	assertParaEnd2(t, 0, "This is a really nice test") // might not be finished
	assertParaEnd2(t, 1, "This is a really nice test, and it has enough tokens\n")
	assertParaEnd2(t, 1, "This is a really nice test, and it has enough tokens\n\n")
	assertParaEnd2(t, 0, "\"This is a really nice test, and it has enough tokens.\"\n\n")
	assertParaEnd2(t, 0, "\"This is a really nice test, and it has enough tokens\"\n")
	assertParaEnd2(t, 0, "\"This is a really nice test, and it has enough tokens\"\n\n")
	assertParaEnd2(t, 0, "This is a test.\n\nRegards,\nJim")
	assertParaEnd2(t, 0, "This is a test.\n\nRegards,\n\nJim")
	assertParaEnd2(t, 0, "This is a test.\n\nKind Regards,\nJim")
	assertParaEnd2(t, 0, "This is a test.\n\nKind Regards,\n\nJim")
	assertParaEnd2(t, 0, "This is a test.\n\nKind Regards,\n\nJim Tester")
	assertParaEnd2(t, 0, "This is a test.\n\nKind Regards,\n\nJim van Tester")
	assertParaEnd2(t, 0, "This is headline-style text")
	assertParaEnd2(t, 0, "This is headline-style text.")
	assertParaEnd2(t, 0, "This is headline-style text. If it gets longer, a dot is needed.")
	assertParaEnd2(t, 1, "This is headline-style text. If it gets longer, a dot is needed")
	assertParaEnd2(t, 0, "This is a test\n\nKind Regards,\n\nJim van Tester")
	assertParaEnd2(t, 1, "This is a really nice test, and it has enough tokens\n\nKind Regards,\n\nJim van Tester")
}
