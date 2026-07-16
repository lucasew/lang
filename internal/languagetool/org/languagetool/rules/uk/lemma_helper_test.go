package uk

import (
	"testing"

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
