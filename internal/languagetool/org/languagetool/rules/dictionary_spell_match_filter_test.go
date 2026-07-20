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

// Java String / Hit.begin/end are UTF-16: non-BMP (emoji) before a phrase shifts
// UTF-16 indices by 2 while Go string bytes are 4. Byte-based phrase hits would
// not cover Java RuleMatch positions (UTF-16), so matches would leak.
func TestDictionarySpellMatchFilter_UTF16Positions(t *testing.T) {
	// "😀 San Juan" — emoji is one rune / 2 UTF-16 units / 4 UTF-8 bytes
	text := "😀 San Juan"
	// UTF-16: [😀=0..2][space=2][San=3..6][space=6][Juan=7..11]
	// byte:   [😀=0..4][space=4][San=5..8][space=8][Juan=9..13]
	f := NewDictionarySpellMatchFilter([]string{"San Juan"})
	// Spell matches on San / Juan at UTF-16 indices (Java RuleMatch positions)
	mSan := NewRuleMatch(&DictFilterRule{ID: "SPELL"}, nil, 3, 6, "misspelled")
	mJuan := NewRuleMatch(&DictFilterRule{ID: "SPELL"}, nil, 7, 11, "misspelled")
	out := f.Filter([]*RuleMatch{mSan, mJuan}, text)
	require.Empty(t, out, "UTF-16 phrase hit must cover Java match positions after non-BMP")

	// GetPhrases uses Java String.substring (UTF-16) for covered terms in keys
	phrases := f.GetPhrases([]*RuleMatch{mSan, mJuan}, text)
	require.Contains(t, phrases, "San Juan")
	require.Len(t, phrases["San Juan"], 2)
}
