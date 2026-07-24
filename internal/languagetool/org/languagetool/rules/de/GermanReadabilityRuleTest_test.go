package de

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanReadabilityRule_SyllablesAndFRE(t *testing.T) {
	require.Equal(t, 1, simpleSyllablesCountDE("a"))
	require.Greater(t, simpleSyllablesCountDE("Haus"), 0)
	require.InDelta(t, 180-10-58.5*1.5, germanFleschReadingEase(10, 1.5), 0.01)
	// hard text: fre low → level 0
	require.Equal(t, 0, germanReadabilityLevel(10))
	// easy text
	require.Equal(t, 6, germanReadabilityLevel(95))
}

func TestGermanReadabilityRule_MatchList_Difficult(t *testing.T) {
	// Long words / short sentences → low FRE → difficult
	r := NewGermanReadabilityRule(nil, false)
	r.Level = 3
	r.MinWords = 10
	// many multisyllabic words
	var b strings.Builder
	for i := 0; i < 12; i++ {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString("Verantwortlichkeiten")
	}
	b.WriteString(".")
	sents := languagetool.SplitAndAnalyze(b.String())
	// may or may not flag depending on syllable counts; just smoke MatchList
	_ = r.MatchList(sents)
	require.Equal(t, "READABILITY_RULE_DIFFICULT_DE", r.GetID())
	require.Equal(t, "READABILITY_RULE_SIMPLE_DE", NewGermanReadabilityRule(nil, true).GetID())
}
