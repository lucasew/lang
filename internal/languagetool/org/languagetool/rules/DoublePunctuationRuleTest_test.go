package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.DoublePunctuationRuleTest

func TestDoublePunctuationRule_Rule(t *testing.T) {
	rule := NewDoublePunctuationRule(map[string]string{
		"two_dots":          "Two consecutive dots",
		"two_commas":        "Two consecutive commas",
		"double_dots_short": "two dots",
		"double_commas_short": "two commas",
		"desc_double_punct": "Double punctuation",
	})

	assert0 := func(s string) {
		t.Helper()
		m := rule.Match(languagetool.AnalyzePlain(s))
		require.Equal(t, 0, len(m), "input %q", s)
	}

	assert0("This is a test sentence...")
	assert0("Это тестовое предложение?..")
	assert0("Это тестовое предложение!.. ")
	assert0("This is a test sentence... More stuff....")
	assert0("This is a test sentence..... More stuff....")
	assert0("This, is, a test sentence.")
	assert0("The path is ../data/vtest.avi")
	assert0("The path is ..\\data\\vtest.avi")
	assert0("Something … … ..")
	assert0("Something … … ... …")
	assert0("Something … … .... …")
	assert0("Something … … .. …")
	assert0("Something ……..")

	matches1 := rule.Match(languagetool.AnalyzePlain("This,, is a test sentence."))
	require.Equal(t, 1, len(matches1))
	require.Equal(t, 4, matches1[0].GetFromPos())
	require.Equal(t, 6, matches1[0].GetToPos())

	matches2 := rule.Match(languagetool.AnalyzePlain("This is a test sentence.. Another sentence"))
	require.Equal(t, 1, len(matches2))
	require.Equal(t, 23, matches2[0].GetFromPos())
	require.Equal(t, 25, matches2[0].GetToPos())
}
