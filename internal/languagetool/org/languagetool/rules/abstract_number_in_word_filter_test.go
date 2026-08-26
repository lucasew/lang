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
		return w != "word" && w != "wrd" && w != "cas" && w != "ok"
	}
	// no digits + misspelled → empty (Java also empty)
	require.Empty(t, f.Suggestions("hello"))
	// no digits + not misspelled → Java still adds wordWithoutNumberCharacter (= word)
	require.Equal(t, []string{"ok"}, f.Suggestions("ok"))
	require.Equal(t, []string{"word", "wrd"}, f.Suggestions("w0rd"))
	require.Equal(t, []string{"cas"}, f.Suggestions("cas4"))
	// digit-only: without → ""; not misspelled "" → Java adds ""
	f.IsMisspelled = func(w string) bool {
		return w != "" && w != "o"
	}
	require.Equal(t, []string{"o", ""}, f.Suggestions("0"))

	f.IsMisspelled = func(w string) bool {
		return w != "word" && w != "wrd" && w != "cas"
	}
	m := NewRuleMatch(nil, nil, 0, 4, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"word": "cas4"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"cas"}, out.GetSuggestedReplacements())
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"word": "hello"}, 0, nil, nil))
}

func TestNumberInWordFilter_GetSuggestionsFallback(t *testing.T) {
	// Java: when both gates fail, call getSuggestions(wordWithoutNumberCharacter)
	f := NewNumberInWordFilter()
	f.IsMisspelled = func(w string) bool { return true }
	f.GetSuggestions = func(w string) []string {
		if w == "tst" {
			return []string{"test", "toast"}
		}
		return nil
	}
	require.Equal(t, []string{"test", "toast"}, f.Suggestions("t5st"))
}
