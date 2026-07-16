package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDictionarySpellMatchFilter(t *testing.T) {
	f := NewDictionarySpellMatchFilter([]string{"San Juan"})
	text := "Visit San Juan today"
	// "San" 6-9, "Juan" 10-14 in "Visit San Juan today"
	m1 := NewRuleMatch(&DictFilterRule{ID: "SPELL"}, nil, 6, 9, "misspelled")
	m2 := NewRuleMatch(&DictFilterRule{ID: "SPELL"}, nil, 10, 14, "misspelled")
	m3 := NewRuleMatch(&DictFilterRule{ID: "SPELL"}, nil, 0, 5, "misspelled") // Visit
	out := f.Filter([]*RuleMatch{m1, m2, m3}, text)
	require.Len(t, out, 1)
	require.Equal(t, 0, out[0].FromPos)

	// non-dict rule kept even inside phrase
	m4 := NewRuleMatch(NewFakeRule("OTHER"), nil, 6, 9, "x")
	out = f.Filter([]*RuleMatch{m4}, text)
	require.Len(t, out, 1)
}
