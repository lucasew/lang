package en

// Twin of languagetool-standalone/src/test/java/org/languagetool/rules/en/EnglishForL2SpeakersFalseFriendRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/stretchr/testify/require"
)

// Port of EnglishForL2SpeakersFalseFriendRuleTest.testRule
func TestEnglishForL2SpeakersFalseFriendRule_Rule(t *testing.T) {
	// Java FakeLanguageModel map preferring "complete" over "achieve"
	lm := ngrams.FuncLanguageModel(func(tokens []string) ngrams.Probability {
		joined := ""
		for i, tok := range tokens {
			if i > 0 {
				joined += " "
			}
			joined += tok
		}
		// Higher scores for contexts with complete
		high := map[string]bool{
			"will complete": true, "complete her": true, "complete her task": true,
			"was completed": true, "was completed .": true, "form was completed": true,
		}
		low := map[string]bool{
			"will achieve": true, "achieve her": true, "achieve her task": true,
			"was achieved": true, "was achieved .": true, "form was achieved": true,
		}
		if high[joined] {
			return ngrams.NewProbabilitySimple(0.9, 1.0)
		}
		if low[joined] {
			return ngrams.NewProbabilitySimple(0.01, 1.0)
		}
		// also score single-token preference used by 3-gram path
		for _, tok := range tokens {
			if tok == "complete" || tok == "completed" {
				return ngrams.NewProbabilitySimple(0.8, 1.0)
			}
			if tok == "achieve" || tok == "achieved" {
				return ngrams.NewProbabilitySimple(0.05, 1.0)
			}
		}
		return ngrams.NewProbabilitySimple(0.1, 1.0)
	})

	fr := NewEnglishForFrenchFalseFriendRuleWithLM(lm)
	require.Equal(t, "EN_FOR_FR_SPEAKERS_FALSE_FRIENDS", fr.GetID())
	require.Equal(t, "fr", fr.MotherTongue)
	require.Equal(t, []string{"confusion_sets_l2_fr.txt"}, fr.GetFilenames())

	// Java loads confusion_sets_l2_fr.txt; inject achieve/complete pair for the twin.
	// (full resource load depends on data path; pair is what FakeLanguageModel exercises)
	fr.SetConfusionPair(rules.NewConfusionPairTokens("achieve", "complete", 10, true))
	// also surface form "achieved" if needed
	// second match uses "achieved" — add pair
	// ConfusionProbabilityRule SetConfusionPair replaces map — need multi-pair index
	// Wire both via WordToPairs
	pair1 := rules.NewConfusionPairTokens("achieve", "complete", 10, true)
	pair2 := rules.NewConfusionPairTokens("achieved", "completed", 10, true)
	fr.WordToPairs = map[string][]*rules.ConfusionPair{
		"achieve":  {pair1},
		"achieved": {pair2},
	}

	// Ensure false-friend message path can load achieve→réaliser from official XML
	ClearL2FalseFriendRuleCache()
	fr.FalseFriendsXML = discoverFalseFriendsXML()
	require.NotEmpty(t, fr.FalseFriendsXML, "official false-friends.xml required for twin")
	// Java isBaseformMatch uses English tagger for "achieved" → lemma "achieve"
	fr.TagWord = func(token string) []languagetool.TokenTag {
		switch token {
		case "achieved":
			return []languagetool.TokenTag{{POS: "VBD", Lemma: "achieve"}}
		case "achieve":
			return []languagetool.TokenTag{{POS: "VB", Lemma: "achieve"}}
		default:
			return nil
		}
	}

	m := fr.Match(languagetool.AnalyzePlain("She will achieve her task."))
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetMessage(), `"achieve" (English) means "réaliser" (French)`)
	require.Contains(t, m[0].GetSuggestedReplacements(), "complete")

	m = fr.Match(languagetool.AnalyzePlain("The code only worked if the form was achieved."))
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetMessage(), `"achieve" (English) means "réaliser" (French)`)
	require.Contains(t, m[0].GetSuggestedReplacements(), "completed")

	// clean sentence — no pair hit
	m = fr.Match(languagetool.AnalyzePlain("She will complete her task."))
	require.Empty(t, m)
}
