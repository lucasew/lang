package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.WordRepeatRuleTest — full-strength asserts.

func TestWordRepeatRule_Test(t *testing.T) {
	rule := NewWordRepeatRule(map[string]string{
		"repetition":            "Word repetition",
		"desc_repetition":       "Word repetition",
		"desc_repetition_short": "repetition",
	})

	assertGood := func(s string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(s))
		require.Equal(t, 0, len(matches), "assertGood %q", s)
	}
	assertBad := func(s string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(s))
		require.Equal(t, 1, len(matches), "assertBad %q", s)
	}

	assertGood("A test")
	assertGood("A test.")
	assertGood("A test...")
	assertGood("1 000 000 years")
	assertGood("010 020 030")
	// thumbs up, green heart, evergreen tree x2 as emoji
	assertGood("👍💚🌲🌲")

	assertBad("A A test")
	assertBad("A a test")
	assertBad("This is is a test")

	// Java createRuleMatch sets shortMessage from desc_repetition_short.
	ms := rule.Match(languagetool.AnalyzePlain("This is is a test"))
	require.Len(t, ms, 1)
	require.Equal(t, "repetition", ms[0].GetShortMessage())
	require.Equal(t, "Word repetition", rule.GetDescription())
	require.Equal(t, 1, rule.EstimateContextForSureMatch())
}

func TestWordRepeatRule_CategoryAndIssueType(t *testing.T) {
	r := NewWordRepeatRule(nil)
	require.NotNil(t, r.GetCategory())
	require.Equal(t, NewCategoryId("MISC"), r.GetCategory().GetID())
	require.Equal(t, ITSDuplication, r.GetLocQualityIssueType())
}

func TestWordRepeatBeginningRule_Category(t *testing.T) {
	r := NewWordRepeatBeginningRule(map[string]string{"desc_repetition_beginning": "Successive sentences beginning with the same word"})
	require.NotNil(t, r.GetCategory())
	require.Equal(t, NewCategoryId("REPETITIONS_STYLE"), r.GetCategory().GetID())
	require.Equal(t, ITSStyle, r.GetLocQualityIssueType())
	require.Equal(t, "Successive sentences beginning with the same word", r.GetDescription())
}

func TestParagraphRepeatBeginningRule_Category(t *testing.T) {
	r := NewParagraphRepeatBeginningRule(nil)
	require.NotNil(t, r.GetCategory())
	require.Equal(t, NewCategoryId("STYLE"), r.GetCategory().GetID())
}
