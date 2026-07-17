package ngrams

// Twin of ConfusionProbabilityRuleTest — full n-gram Match deferred; helpers green.
import (
	"testing"

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
}

// Port of ConfusionProbabilityRuleTest.testLocalException
func TestConfusionProbabilityRule_LocalException(t *testing.T) {
	r := NewConfusionProbabilityRule(nil, 3)
	r.Exceptions = []string{"their are new ideas"}
	require.True(t, r.IsLocalException("And Their are new ideas to explore."))
	require.False(t, r.IsLocalException("Their car is broken."))
}
