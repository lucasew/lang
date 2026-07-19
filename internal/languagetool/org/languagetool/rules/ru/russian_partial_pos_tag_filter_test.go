package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestNoDisambiguationRussianPartialPosTagFilter_InjectedTag(t *testing.T) {
	f := NewNoDisambiguationRussianPartialPosTagFilter(func(p string) []string {
		if p == "дом" {
			return []string{"NN:masc"}
		}
		return nil
	})
	ok, err := f.Accept("домов", "^(дом)ов$", "NN.*", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestNoDisambiguationRussianPartialPosTagFilter_FailClosedWithoutTagger(t *testing.T) {
	ClearDefaultRussianPartialPosTagger()
	f := NewNoDisambiguationRussianPartialPosTagFilter(nil)
	ok, err := f.Accept("домов", "^(дом)ов$", "NN.*", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestNoDisambiguationRussianPartialPosTagFilter_DefaultTaggerHook(t *testing.T) {
	SetDefaultRussianPartialPosTagger(func(p string) []string {
		if p == "дом" {
			return []string{"NN:masc"}
		}
		return nil
	})
	t.Cleanup(ClearDefaultRussianPartialPosTagger)

	f := NewNoDisambiguationRussianPartialPosTagFilter(nil)
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("домов", "NN:masc", "дом", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(дом)ов$", "postag_regexp": "NN.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.NotNil(t, out)
}

func TestRussianPartialPosTagFilters_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ru.NoDisambiguationRussianPartialPosTagFilter"))
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ru.RussianPartialPosTagFilter"))
}
