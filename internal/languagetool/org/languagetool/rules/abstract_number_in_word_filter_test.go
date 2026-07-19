package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNumberInWordFilter(t *testing.T) {
	f := NewNumberInWordFilter()
	// without speller: fail-closed
	require.Empty(t, f.Suggestions("w0rd"))
	require.Empty(t, f.Suggestions("cas4"))

	f.IsMisspelled = func(w string) bool {
		return w != "word" && w != "wrd" && w != "cas"
	}
	require.Nil(t, f.Suggestions("hello"))
	require.Equal(t, []string{"word", "wrd"}, f.Suggestions("w0rd"))
	require.Equal(t, []string{"cas"}, f.Suggestions("cas4"))

	m := NewRuleMatch(nil, nil, 0, 4, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"word": "cas4"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"cas"}, out.GetSuggestedReplacements())
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"word": "hello"}, 0, nil, nil))
}
