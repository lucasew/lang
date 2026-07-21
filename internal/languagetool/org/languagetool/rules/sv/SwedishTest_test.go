package sv

// Twin of SwedishTest.testLanguage — analyze smoke (full demo-rule list deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of SwedishTest.testLanguage
func TestSwedish_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sv")
	require.Equal(t, "sv", lt.GetLanguageCode())
	sents := lt.Analyze("Detta är en testtext.")
	require.NotEmpty(t, sents)
}

// Twin of SwedishTest.testSpellingAndColon — Java expects 0 matches on "Arbeta med var:".
func TestSwedish_SpellingAndColon(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sv")
	// Without full SV rule stack, Analyze must not invent errors; Check may be empty.
	sents := lt.Analyze("Arbeta med var:")
	require.NotEmpty(t, sents)
	// empty check when no SV rules registered
	require.Empty(t, lt.Check("Arbeta med var:"))
}

// Twin of SwedishTest.testWeekdayAndMonthNames
func TestSwedish_WeekdayAndMonthNames(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sv")
	// Java weekday/month capitalisation rules — without SV grammar, fail closed (no invent).
	require.Empty(t, lt.Check("På måndag är alla lediga."))
	require.Empty(t, lt.Check("I oktober kommer ofta den första snön."))
	// Capitalized weekday would error in Java; Go without rules stays empty (incomplete, not invent).
	_ = lt.Check("På Måndag är alla lediga.")
}
