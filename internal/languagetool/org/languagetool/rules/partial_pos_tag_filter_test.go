package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPartialPosTagFilter(t *testing.T) {
	f := NewPartialPosTagFilter(func(partial string) []string {
		if partial == "happy" {
			return []string{"JJ"}
		}
		return []string{"NN"}
	})
	ok, err := f.Accept("unhappy", "un(.*)", "JJ", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = f.Accept("unhappy", "un(.*)", "VB", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = f.Accept("unhappy", "un(.*)", "JJ", true, false, "", "")
	require.NoError(t, err)
	require.False(t, ok) // negated: has JJ → false
}

func TestPartialPosTagFilter_AcceptRuleMatch(t *testing.T) {
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
	require.NotNil(t, out)

	// fail-closed without tagger
	require.Nil(t, NewPartialPosTagFilter(nil).AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "(.*)", "postag_regexp": "JJ",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil))
}
