package ru

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Java rule demo sentences (addExamplePair) — correction = fixed marker span.
func TestRU_ExamplePairs(t *testing.T) {
	require.Equal(t, []string{"Эспрессо"}, NewRussianSimpleReplaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"доме"}, NewRussianWordRepeatRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"конференц-зале"}, NewRussianCompoundRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"оффлайн"}, NewRussianWordCoherencyRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"("}, NewRussianUnpairedBracketsRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Рытый Банк"}, NewRussianSpecificCaseRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Я иду"}, NewRussianVerbConjugationRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"абрикосов"}, NewRussianWordRootRepeatRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"каждая"}, NewMorfologikRussianSpellerRule().GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"каждая"}, NewMorfologikRussianYOSpellerRule().GetIncorrectExamples()[0].GetCorrections())
}
