package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestRussianPartialPosTagFilters_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ru.NoDisambiguationRussianPartialPosTagFilter"))
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ru.RussianPartialPosTagFilter"))
}

func TestRussianPartialPosTagFilter_FailClosedWithoutHooks(t *testing.T) {
	ClearDefaultRussianPartialPosTagger()
	ClearRussianFilterDisambiguator()
	f := NewRussianPartialPosTagFilter(nil)
	ok, err := f.Accept("дом", "^(дом)$", "NN.*", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestRussianPartialPosTagFilter_WithTagAndDisambig(t *testing.T) {
	SetDefaultRussianPartialPosTagger(func(p string) []string {
		if p == "дом" {
			return []string{"NN:masc", "extra"}
		}
		return nil
	})
	// Identity disambiguator keeps tags.
	WireRussianFilterDisambiguator(stubRUDisambig{})
	t.Cleanup(func() {
		ClearDefaultRussianPartialPosTagger()
		ClearRussianFilterDisambiguator()
	})
	f := NewRussianPartialPosTagFilter(nil)
	ok, err := f.Accept("домов", "^(дом)ов$", "NN.*", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

type stubRUDisambig struct{}

func (stubRUDisambig) Disambiguate(s *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return s
}
