package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishSpecificCaseRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishSpecificCaseRule_Rule(t *testing.T) {
	rule := NewEnglishSpecificCaseRule(nil)
	assertGood := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(s))), "good %q", s)
	}

	assertGood("Harry Potter")
	assertGood("I like Harry Potter.")
	assertGood("I like HARRY POTTER.")

	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("harry potter"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("harry Potter"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Harry potter"))))

	matches1 := rule.Match(languagetool.AnalyzePlain("I like Harry potter."))
	require.Equal(t, 1, len(matches1))
	require.Equal(t, 7, matches1[0].GetFromPos())
	require.Equal(t, 19, matches1[0].GetToPos())
	require.Equal(t, []string{"Harry Potter"}, matches1[0].GetSuggestedReplacements())
	require.Equal(t, "If the term is a proper noun, use initial capitals.", matches1[0].GetMessage())

	matches2 := rule.Match(languagetool.AnalyzePlain("Alexander The Great"))
	require.Equal(t, 1, len(matches2))
	require.Equal(t, "If the term is a proper noun, use the suggested capitalization.", matches2[0].GetMessage())

	matches3 := rule.Match(languagetool.AnalyzePlain("I like Harry  potter."))
	require.Equal(t, 1, len(matches3))
	require.Equal(t, 7, matches3[0].GetFromPos())
	require.Equal(t, 20, matches3[0].GetToPos())
}
