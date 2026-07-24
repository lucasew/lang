package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPartialPosTagFilter_AcceptMatrix(t *testing.T) {
	f := NewPartialPosTagFilter(func(partial string) []string {
		if partial == "happy" {
			return []string{"JJ"}
		}
		if partial == "run" {
			return []string{"VB", "NN"}
		}
		return []string{"NN"}
	})
	ok, err := f.Accept("unhappy", "un(.*)", "JJ", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = f.Accept("unhappy", "un(.*)", "VB", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)

	// negate_pos: has JJ → false
	ok, err = f.Accept("unhappy", "un(.*)", "JJ", true, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)

	// multi-reading: any match keeps (non-negate)
	ok, err = f.Accept("rerun", "re(.*)", "VB", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)

	// two groups concatenate
	f2 := NewPartialPosTagFilter(func(partial string) []string {
		if partial == "ab" {
			return []string{"X"}
		}
		return nil
	})
	ok, err = f2.Accept("ab", "(a)(b)", "X", false, true, "", "")
	require.NoError(t, err)
	require.True(t, ok)

	// prefix/suffix wrap surface before matcher.matches (Java prefix+token+suffix).
	// token "b" + prefix "a" + suffix "" → "ab" with groups (a)(b).
	ok, err = f2.Accept("b", "(a)(b)", "X", false, true, "a", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPartialPosTagFilter_AcceptRuleMatchIdentityAndDrop(t *testing.T) {
	f := NewPartialPosTagFilter(func(partial string) []string {
		if partial == "accurate" {
			return []string{"JJ"}
		}
		return nil
	})
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("inaccurate", nil, nil), 0)
	m := NewRuleMatch(nil, nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "(?:in|un)(.*)", "postag_regexp": "JJ",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.Equal(t, m, out, "Java returns same match object")

	// POS miss → null
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "(?:in|un)(.*)", "postag_regexp": "VB",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil))

	// fail-closed without tagger
	require.Nil(t, NewPartialPosTagFilter(nil).AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "(.*)", "postag_regexp": "JJ",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil))
}

func TestPartialPosTagFilter_MissingArgsPanic(t *testing.T) {
	f := NewPartialPosTagFilter(func(partial string) []string { return []string{"JJ"} })
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("x", nil, nil), 0)
	m := NewRuleMatch(nil, nil, 0, 1, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{"no": "1", "regexp": "(.*)"}, 0,
			[]*languagetool.AnalyzedTokenReadings{tok}, nil)
	})
}

func TestPartialPosTagFilter_WrongGroupCountPanic(t *testing.T) {
	f := NewPartialPosTagFilter(func(partial string) []string { return []string{"JJ"} })
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("x", nil, nil), 0)
	m := NewRuleMatch(nil, nil, 0, 1, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "regexp": "x", "postag_regexp": "JJ",
		}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	})
}

func TestPartialPosTagFilter_BadIndexPanic(t *testing.T) {
	f := NewPartialPosTagFilter(func(partial string) []string { return []string{"JJ"} })
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("x", nil, nil), 0)
	m := NewRuleMatch(nil, nil, 0, 1, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "2", "regexp": "(.*)", "postag_regexp": "JJ",
		}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	})
}
