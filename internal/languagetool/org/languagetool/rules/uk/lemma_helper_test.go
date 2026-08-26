package uk

import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestLemmaHelper_Sets(t *testing.T) {
	require.Contains(t, MonthLemmas, "січень")
	require.Contains(t, DaysOfWeek, "понеділок")
	require.True(t, IsTimePlusLemma("рік"))
	require.True(t, IsTimePlusLemma("кілометр"))
	require.True(t, HasLemmaInList([]string{"рік", "foo"}, TimeLemmasShort))
	require.False(t, HasLemmaString([]string{"а"}, "б"))
	require.Equal(t, "тест", CleanIgnoreChars("те\u0301ст"))
}

func TestLemmaHelper_Patterns(t *testing.T) {
	require.True(t, AdvQuantPattern.MatchString("багато"))
	require.True(t, PartInsertPattern.MatchString("навіть"))
	require.True(t, QuotesPattern.MatchString("«"))
}

func TestReverseForwardSearchIdx(t *testing.T) {
	zd := "здатний"
	tokens := []*languagetool.AnalyzedTokenReadings{
		atr("X"),
		atrLemma("здатні", &zd, "adj:p:v_naz"),
		atr("громадяни", "noun:anim:p:v_naz"),
		atr("вважатися", "verb:imperf:inf"),
	}
	// reverse from noun-1 (pos 1) looking for infAgreement lemma
	idx := ReverseSearchIdx(tokens, 1, 6, infAgreementPattern, nil)
	require.Equal(t, 1, idx)
	// forward from after verb-start
	idx2 := ForwardLemmaSearchIdx(tokens, 0, 5, infAgreementPattern, nil)
	require.Equal(t, 1, idx2)
	// Java matches(): pattern must cover full POS tag
	require.True(t, ReverseSearch(tokens, 2, 3, nil, regexp.MustCompile(`^adj.*`)))
	require.True(t, ForwardPosTagSearch(tokens, 1, "verb", 3))
}
