package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestEnglishPartialPosTagFilter(t *testing.T) {
	f := NewNoDisambiguationEnglishPartialPosTagFilter(func(p string) []string {
		if p == "happy" {
			return []string{"JJ"}
		}
		return nil
	})
	ok, err := f.Accept("unhappy", "un(.*)", "JJ", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestNoDisambiguationEnglishPartialPosTagFilter_FailClosedWithoutTagger(t *testing.T) {
	ClearEnglishFilterTagger()
	f := NewNoDisambiguationEnglishPartialPosTagFilter(nil)
	// default tag uses process-wide tagger; none wired → no POS → drop
	ok, err := f.Accept("unhappy", "un(.*)", "JJ", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)

	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "JJ", "happy", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "un(.*)", "postag_regexp": "JJ",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
}

func TestNoDisambiguationEnglishPartialPosTagFilter_AcceptRuleMatch(t *testing.T) {
	f := NewNoDisambiguationEnglishPartialPosTagFilter(func(p string) []string {
		if p == "happy" {
			return []string{"JJ"}
		}
		return nil
	})
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "JJ", "happy", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "un(.*)", "postag_regexp": "JJ",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.NotNil(t, out)
}

func TestNoDisambiguationEnglishPartialPosTagFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.en.NoDisambiguationEnglishPartialPosTagFilter"))
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.en.EnglishPartialPosTagFilter"))
}
