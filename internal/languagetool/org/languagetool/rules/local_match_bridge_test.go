package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestFromLocalMatches_PreservesMeta(t *testing.T) {
	sent := languagetool.AnalyzePlain("A gift.")
	ms := []languagetool.LocalMatch{{
		RuleID:       "GIFT",
		FromPos:      2,
		ToPos:        6,
		Message:      "false friend",
		ShortMessage: "ff",
		Suggestions:  []string{"Geschenk"},
		CategoryID:   "FALSEFRIENDS",
		CategoryName: "False Friends",
		IssueType:    "misspelling",
	}}
	out := FromLocalMatches(ms, sent)
	require.Len(t, out, 1)
	require.Equal(t, "misspelling", out[0].IssueType)
	require.Equal(t, "FALSEFRIENDS", out[0].CategoryID)
	require.Equal(t, "False Friends", out[0].CategoryName)
	require.Equal(t, "ff", out[0].ShortMessage)
	require.Equal(t, []string{"Geschenk"}, out[0].GetSuggestedReplacements())
}
