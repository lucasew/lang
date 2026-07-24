package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestNoDisambiguationPortuguesePartialPosTagFilter_InjectedTag(t *testing.T) {
	f := NewNoDisambiguationPortuguesePartialPosTagFilter(func(p string) []string {
		if p == "casa" {
			return []string{"NCFS000"}
		}
		return nil
	})
	ok, err := f.Accept("casas", "^(casa)s$", "NC.*", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestNoDisambiguationPortuguesePartialPosTagFilter_FailClosedWithoutTagger(t *testing.T) {
	ClearDefaultPortuguesePartialPosTagger()
	f := NewNoDisambiguationPortuguesePartialPosTagFilter(nil)
	ok, err := f.Accept("casas", "^(casa)s$", "NC.*", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestNoDisambiguationPortuguesePartialPosTagFilter_DefaultTaggerHook(t *testing.T) {
	SetDefaultPortuguesePartialPosTagger(func(p string) []string {
		if p == "casa" {
			return []string{"NCFS000"}
		}
		return nil
	})
	t.Cleanup(ClearDefaultPortuguesePartialPosTagger)

	f := NewNoDisambiguationPortuguesePartialPosTagFilter(nil)
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("casas", "NCFP000", "casa", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(casa)s$", "postag_regexp": "NC.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.NotNil(t, out)
}

func TestNoDisambiguationPortuguesePartialPosTagFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.pt.NoDisambiguationPortuguesePartialPosTagFilter"))
}
