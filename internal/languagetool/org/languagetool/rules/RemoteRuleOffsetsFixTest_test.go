package rules

// Twin of RemoteRuleOffsetsFixTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func printShifts(text string) []int {
	return ComputeOffsetShifts(text)
}

// Port of RemoteRuleOffsetsFixTest.testShiftCalculation
func TestRemoteRuleOffsetsFix_ShiftCalculation(t *testing.T) {
	require.Equal(t, []int{0, 2, 3, 4, 5, 6}, printShifts("😁foo"))
	require.Equal(t, []int{0, 1, 2, 3, 4, 6, 7, 8, 9, 10, 11}, printShifts("foo 😁 bar"))
	require.Equal(t, []int{0, 2, 3, 4, 5, 6, 7, 9, 10, 11, 12, 13, 14, 15}, printShifts("😁 foo 😁 bar"))
	require.Equal(t, []int{0, 2, 3}, printShifts("👪"))
	require.Equal(t, []int{0, 2, 4, 5, 6}, printShifts("👍🏼"))
	require.Equal(t, []int{0, 1}, printShifts("a"))
}

func sentenceWithText(text string) *languagetool.AnalyzedSentence {
	tok := languagetool.NewAnalyzedToken(text, nil, nil)
	return languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(tok, 0),
	})
}

// Port of RemoteRuleOffsetsFixTest.testMatches
func TestRemoteRuleOffsetsFix_Matches(t *testing.T) {
	s := sentenceWithText("😁foo bar")
	r := NewFakeRule("FAKE")
	matches := []*RuleMatch{
		NewRuleMatch(r, s, 0, 1, "Emoji"),
		NewRuleMatch(r, s, 1, 4, "foo"),
	}
	FixMatchOffsets(s, matches)
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 2, matches[0].GetToPos())
	require.Equal(t, 2, matches[1].GetFromPos())
	require.Equal(t, 5, matches[1].GetToPos())
}

// Port of RemoteRuleOffsetsFixTest.testException — nil-safe
func TestRemoteRuleOffsetsFix_Exception(t *testing.T) {
	FixMatchOffsets(nil, nil)
	FixMatchOffsets(sentenceWithText("ok"), nil)
}
