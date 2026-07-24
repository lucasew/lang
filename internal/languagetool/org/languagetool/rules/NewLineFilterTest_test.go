package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.NewLineFilterTest (NewLineMatchFilter).

func TestNewLineFilter_Filter(t *testing.T) {
	filter := NewNewLineMatchFilter()
	text1 := "I’m+the+Heeadline\n\n\n\n\u2063\n\n\n\nI’m+some+plain+teext."
	// note: Java uses curly apostrophe ’ (U+2019)
	runNewLineCase(t, filter, text1, 8, 21, "Headline\n\n\n\n\u2063\n\n\n\n", 8, 17, true, "Headline")
	runNewLineCase(t, filter, text1, 8, 21, "Heeadline\n\n\n\n\u2063\n\n\n\n", 8, 17, false, "")
}

func runNewLineCase(t *testing.T, filter *NewLineMatchFilter, text string,
	fromBefore, toBefore int, suggestionBefore string,
	fromAfter, toAfter int, expectMatch bool, suggestionAfter string) {
	t.Helper()
	sentence := languagetool.AnalyzePlain(text)
	rm := NewRuleMatch(NewFakeRule(""), sentence, fromBefore, toBefore, "match1")
	rm.SetSuggestedReplacement(suggestionBefore)
	filtered := filter.Filter([]*RuleMatch{rm}, text)
	if expectMatch {
		require.Len(t, filtered, 1)
		m := filtered[0]
		require.Equal(t, fromAfter, m.GetFromPos())
		require.Equal(t, toAfter, m.GetToPos())
		require.Contains(t, m.GetSuggestedReplacements(), suggestionAfter)
	} else {
		require.Empty(t, filtered)
	}
}
