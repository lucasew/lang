package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSoftRuleMeta(t *testing.T) {
	id, name, issue, short := SoftRuleMeta("EN_A_VS_AN")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "Grammar", name)
	require.Equal(t, "grammar", issue)
	require.Equal(t, "Wrong article", short)

	id, _, issue, _ = SoftRuleMeta("MORFOLOGIK_RULE_EN_US")
	require.Equal(t, "TYPOS", id)
	require.Equal(t, "misspelling", issue)

	id, _, _, _ = SoftRuleMeta("WHITESPACE_RULE")
	require.Equal(t, "TYPOGRAPHY", id)

	// Soft grammar IDs must not be classified as false friends.
	id, name, issue, _ = SoftRuleMeta("EN_SOFT_YOUR_YOU_RE")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "Grammar", name)
	require.Equal(t, "grammar", issue)

	id, _, issue, _ = SoftRuleMeta("FALSE_FRIEND_RULE")
	require.Equal(t, "FALSEFRIENDS", id)
	require.Equal(t, "misspelling", issue)
}

func TestSoftRuleDescription(t *testing.T) {
	require.Equal(t, "Use of 'a' versus 'an'", SoftRuleDescription("EN_A_VS_AN"))
	require.Equal(t, "YOUR YOU RE", SoftRuleDescription("EN_SOFT_YOUR_YOU_RE"))
	require.Equal(t, "Word repetition", SoftRuleDescription("ENGLISH_WORD_REPEAT_RULE"))
}

func TestApiV2_MatchCategory(t *testing.T) {
	api := NewApiV2(nil, nil)
	r, err := api.Handle("check", map[string]string{
		"language": "en",
		"text":     "This is an test.",
	})
	require.NoError(t, err)
	require.Contains(t, r.Body, `"id":"GRAMMAR"`)
	require.Contains(t, r.Body, "Wrong article")
	require.Contains(t, r.Body, "shortMessage")
	require.Contains(t, r.Body, "Use of 'a' versus 'an'")
}
