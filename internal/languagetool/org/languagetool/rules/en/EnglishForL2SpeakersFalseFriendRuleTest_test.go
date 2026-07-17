package en

// Twin of languagetool-standalone/src/test/java/org/languagetool/rules/en/EnglishForL2SpeakersFalseFriendRuleTest.java
// Full ngram FakeLanguageModel deferred — inject Pairs + message surface.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of EnglishForL2SpeakersFalseFriendRuleTest.testRule
func TestEnglishForL2SpeakersFalseFriendRule_Rule(t *testing.T) {
	fr := NewEnglishForFrenchFalseFriendRule()
	require.Equal(t, "EN_FOR_FR_SPEAKERS_FALSE_FRIENDS", fr.GetID())
	require.Equal(t, "fr", fr.MotherTongue)
	require.Equal(t, []string{"confusion_sets_l2_fr.txt"}, fr.GetFilenames())

	// Java: achieve → complete with FR mother tongue message containing réaliser
	fr.Pairs = []L2ConfusionPair{
		{Wrong: "achieve", Better: "complete", MotherGloss: "réaliser"},
		{Wrong: "achieved", Better: "completed", MotherGloss: "réaliser", MessageWord: "achieve"},
	}

	m, err := fr.Match(languagetool.AnalyzePlain("She will achieve her task."))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetMessage(), `"achieve" (English) means "réaliser" (French)`)
	require.Contains(t, m[0].GetSuggestedReplacements(), "complete")

	m, err = fr.Match(languagetool.AnalyzePlain("The code only worked if the form was achieved."))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetMessage(), `"achieve" (English) means "réaliser" (French)`)

	// clean sentence — no pair hit
	m, err = fr.Match(languagetool.AnalyzePlain("She will complete her task."))
	require.NoError(t, err)
	require.Empty(t, m)
}
