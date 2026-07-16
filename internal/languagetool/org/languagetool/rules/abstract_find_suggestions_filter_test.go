package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAbstractFindSuggestionsFilter(t *testing.T) {
	f := &AbstractFindSuggestionsFilter{
		SpellingSuggestions: func(atr *languagetool.AnalyzedTokenReadings) []string {
			return []string{"café", "cafe", "boat"}
		},
		MatchesDesiredPostag: func(suggestion, postag string) bool {
			return suggestion != "boat"
		},
	}
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("cafe", nil, nil))
	tok.SetStartPos(0)
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 4, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*", "Mode": "diacritics",
	}, []*languagetool.AnalyzedTokenReadings{tok}, []int{1})
	require.NotNil(t, out)
	require.Equal(t, []string{"café"}, out.GetSuggestedReplacements())
}
