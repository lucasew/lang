package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.LongParagraphRuleTest

func TestLongParagraphRule_Rule(t *testing.T) {
	rule := NewLongParagraphRule(map[string]string{
		"long_paragraph_rule_msg":  "This paragraph is too long (%d words)",
		"long_paragraph_rule_desc": "Paragraph over %d words",
	}, 6)

	assertLen := func(text string, n int) {
		t.Helper()
		m := rule.MatchList(languagetool.AnalyzeTextDemo(text))
		require.Equal(t, n, len(m), "text=%q", text)
	}

	assertLen("This is a short paragraph.", 0)
	assertLen("This is only almost long paragraph by unit test standards.", 0)
	assertLen("Here's some text as a filler. This is a long paragraph by unit test standards.", 1)
	assertLen("Here's some text as a filler.  A test. A long paragraph by unit test standards.", 1)
	assertLen("Here's some text as a filler.  A test.\nNot a long paragraph.\nBecause of the line breaks.\n", 0)
	assertLen("- [ ] A test.\n- [ ] Not a long paragraph.\n- [ ] Because of the line breaks.\n- [ ] More text even.\n", 0)

	text1 := "This is a short paragraph.\n\nHere's some text as filler. This is a long paragraph by unit test standards."
	matches1 := rule.MatchList(languagetool.AnalyzeTextDemo(text1))
	require.Equal(t, 1, len(matches1))
	require.Equal(t, 45, matches1[0].GetFromPos())
	require.Equal(t, 54, matches1[0].GetToPos())

	text2 := "Here's some text as filler. This is a long paragraph by unit test standards.\n\nAnother paragraph.\n\nHere's some text as morefiller - this is a long paragraph by unit test standards."
	matches2 := rule.MatchList(languagetool.AnalyzeTextDemo(text2))
	require.Equal(t, 2, len(matches2))
	require.Equal(t, 17, matches2[0].GetFromPos())
	require.Equal(t, 26, matches2[0].GetToPos())
	require.Equal(t, 115, matches2[1].GetFromPos())
	require.Equal(t, 128, matches2[1].GetToPos())
}
