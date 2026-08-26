package spelling

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/suggestions"
	"github.com/stretchr/testify/require"
)

// Twin of SpellingCheckRule.addSuggestionsToRuleMatch — no-reranking arm.
func TestAddSuggestionsToRuleMatch_NoRerank(t *testing.T) {
	sent := languagetool.AnalyzePlain("teh cat")
	m := rules.NewRuleMatch(rules.NewFakeRule("SPELL"), sent, 0, 3, "misspell")
	// pre-existing suggestion on match
	m.SetSuggestedReplacementObjects([]*rules.SuggestedReplacement{
		rules.NewSuggestedReplacement("pre"),
	})
	user := rules.ConvertSuggestions([]string{"teh_user"})
	def := rules.ConvertSuggestions([]string{"the", "tea"})
	AddSuggestionsToRuleMatch("teh", user, def, nil, m)
	got := m.GetSuggestedReplacements()
	require.Equal(t, []string{"pre", "teh_user", "the", "tea"}, got)
}

// Identity orderer is not ML → same concat path.
func TestAddSuggestionsToRuleMatch_IdentityOrdererNotML(t *testing.T) {
	sent := languagetool.AnalyzePlain("teh")
	m := rules.NewRuleMatch(rules.NewFakeRule("SPELL"), sent, 0, 3, "misspell")
	AddSuggestionsToRuleMatchStrings("teh", []string{"u"}, []string{"the"}, suggestions.IdentitySuggestionsOrderer{}, m)
	require.Equal(t, []string{"u", "the"}, m.GetSuggestedReplacements())
}

// SuggestionsRanker path: user candidates prepended; autoCorrect false with user.
func TestAddSuggestionsToRuleMatch_RankerWithUser(t *testing.T) {
	sent := languagetool.AnalyzePlain("teh")
	m := rules.NewRuleMatch(rules.NewFakeRule("SPELL"), sent, 0, 3, "misspell")
	// Ranker that claims ML available and reverses candidates for visibility
	ranker := &mockRanker{ml: true, minConf: 0.9}
	user := rules.ConvertSuggestions([]string{"myteh"})
	def := rules.ConvertSuggestions([]string{"the", "tea"})
	AddSuggestionsToRuleMatch("teh", user, def, ranker, m)
	got := m.GetSuggestedReplacements()
	require.Equal(t, "myteh", got[0], "user first: %v", got)
	require.Contains(t, got, "the")
	require.False(t, m.GetAutoCorrect(), "no autoCorrect with user dict")
}

// SuggestionsRanker path without user: shouldAutoCorrect from ranker.
func TestAddSuggestionsToRuleMatch_RankerAutoCorrect(t *testing.T) {
	sent := languagetool.AnalyzePlain("teh")
	m := rules.NewRuleMatch(rules.NewFakeRule("SPELL"), sent, 0, 3, "misspell")
	c := float32(0.99)
	ranker := &mockRanker{ml: true, minConf: 0.9, forceConf: &c}
	AddSuggestionsToRuleMatch("teh", nil, rules.ConvertSuggestions([]string{"the", "tea"}), ranker, m)
	require.True(t, m.GetAutoCorrect())
	require.Equal(t, "the", m.GetSuggestedReplacements()[0])
}

// FeatureExtractor rejects user candidates (Java IllegalStateException).
func TestAddSuggestionsToRuleMatch_FeatureExtractorPanicsOnUser(t *testing.T) {
	sent := languagetool.AnalyzePlain("teh")
	m := rules.NewRuleMatch(rules.NewFakeRule("SPELL"), sent, 0, 3, "misspell")
	fe := suggestions.NewSuggestionsOrdererFeatureExtractor(nil)
	// Force ML available by setting a dummy LM
	fe.LM = &mockLM{}
	fe.Score = "noop"
	require.Panics(t, func() {
		AddSuggestionsToRuleMatch("teh",
			rules.ConvertSuggestions([]string{"u"}),
			rules.ConvertSuggestions([]string{"the"}),
			fe, m)
	})
}

// FeatureExtractor without user sets features map (Java candidateCount + per-sug features).
func TestAddSuggestionsToRuleMatch_FeatureExtractor(t *testing.T) {
	sent := languagetool.AnalyzePlain("teh")
	m := rules.NewRuleMatch(rules.NewFakeRule("SPELL"), sent, 0, 3, "misspell")
	fe := suggestions.NewSuggestionsOrdererFeatureExtractor(&mockLM{})
	fe.Score = "noop" // avoid unknown-score panic when no experiment
	AddSuggestionsToRuleMatch("teh", nil, rules.ConvertSuggestions([]string{"the", "tea"}), fe, m)
	require.NotEmpty(t, m.GetSuggestedReplacements())
	feats := m.GetFeatures()
	require.Equal(t, float32(2), feats["candidateCount"])
	require.Contains(t, m.GetSuggestedReplacementObjects()[0].GetFeatures(), "levensthein")
}

// mockRanker implements SuggestionsRanker with ML available.
type mockRanker struct {
	ml        bool
	minConf   float32
	forceConf *float32
}

func (m *mockRanker) IsMlAvailable() bool { return m != nil && m.ml }

func (m *mockRanker) OrderSuggestions(suggestions []string, word string, sentence *languagetool.AnalyzedSentence, startPos int) []*rules.SuggestedReplacement {
	out := make([]*rules.SuggestedReplacement, 0, len(suggestions))
	for i, s := range suggestions {
		sr := rules.NewSuggestedReplacement(s)
		if i == 0 && m.forceConf != nil {
			c := *m.forceConf
			sr.SetConfidence(&c)
		}
		out = append(out, sr)
	}
	return out
}

func (m *mockRanker) ShouldAutoCorrect(ranked []*rules.SuggestedReplacement) bool {
	if len(ranked) == 0 || ranked[0] == nil || ranked[0].Confidence == nil {
		return false
	}
	return *ranked[0].Confidence >= m.minConf
}

type mockLM struct{}

func (mockLM) PseudoProbability(tokens []string) float64 { return 0.1 }
func (mockLM) Count(word string) int64                   { return 1 }

var _ suggestions.SuggestionsRanker = (*mockRanker)(nil)
