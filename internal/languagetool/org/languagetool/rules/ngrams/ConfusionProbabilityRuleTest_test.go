package ngrams

// Twin of ConfusionProbabilityRuleTest — Match when LM + pairs present; helpers green.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Port of ConfusionProbabilityRuleTest.testRule (construction + exception surface)
func TestConfusionProbabilityRule_Rule(t *testing.T) {
	lm := UniformLanguageModel(0.1, 1.0)
	r := NewConfusionProbabilityRule(lm, 3)
	require.Equal(t, ConfusionRuleID, r.GetID())
	require.Equal(t, 3, r.Grams)
	// pairs map
	pair := &rules.ConfusionPair{}
	r.SetWordToPairs(map[string][]*rules.ConfusionPair{
		"their": {pair},
		"there": {pair},
	})
	require.Len(t, r.PairsFor("Their"), 1)
	require.Len(t, r.PairsFor("there"), 1)
	require.Nil(t, r.PairsFor("xyz"))
	require.Equal(t, 3, r.EstimateContextForSureMatch())
	require.Equal(t, "Possible word confusion", r.shortDesc())
}

// Java SpecificIdRule id = getId()+"_"+cleanId(term1)+"_"+cleanId(term2);
// shortDesc = statistics_suggest_short_desc; pair description from statistics_rule_description.
func TestConfusionProbabilityRule_SpecificIdRule(t *testing.T) {
	r := NewConfusionProbabilityRule(nil, 3)
	desc := r.pairDescription("there", "their")
	require.Contains(t, desc, "there")
	require.Contains(t, desc, "their")
	id := r.GetID() + "_" + cleanConfusionID("there") + "_" + cleanConfusionID("their")
	require.Equal(t, "CONFUSION_RULE_THERE_THEIR", id)
	idRule := rules.NewSpecificIdRule(id, desc, false, r.Category, r.IssueType, nil)
	require.Equal(t, id, idRule.GetID())
	// Nil LM → Match empty (fail-closed; no invent scores).
	require.Empty(t, r.Match(languagetool.AnalyzePlain("there is a house")))
	// DE-style messages
	r.Messages = map[string]string{
		"statistics_rule_description":   "Mögliche Verwechselungen zwischen ''{0}'' und ''{1}'' erkennen",
		"statistics_suggest_short_desc": "Mögliche Wortverwechselung",
	}
	require.Equal(t, "Mögliche Wortverwechselung", r.shortDesc())
	require.Contains(t, r.pairDescription("seit", "seid"), "seit")
}

// Port of ConfusionProbabilityRuleTest.testLocalException
func TestConfusionProbabilityRule_LocalException(t *testing.T) {
	r := NewConfusionProbabilityRule(nil, 3)
	r.Exceptions = []string{"their are new ideas"}
	require.True(t, r.IsLocalException("And Their are new ideas to explore."))
	require.False(t, r.IsLocalException("Their car is broken."))
}
