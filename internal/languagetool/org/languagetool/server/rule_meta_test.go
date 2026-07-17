package server

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
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

func TestApiV2_DisabledCategories(t *testing.T) {
	api := NewApiV2(nil, nil)
	// a/an is GRAMMAR — disabling GRAMMAR should suppress it
	r, err := api.Handle("check", map[string]string{
		"language":           "en",
		"text":               "This is an test.",
		"disabledCategories": "GRAMMAR",
	})
	require.NoError(t, err)
	require.NotContains(t, r.Body, "EN_A_VS_AN")
	require.Contains(t, r.Body, `"ignoreRanges"`)
}

func TestApiV2_EnabledCategoriesOnly(t *testing.T) {
	api := NewApiV2(nil, nil)
	r, err := api.Handle("check", map[string]string{
		"language":          "en",
		"text":              "This is an test.",
		"enabledOnly":       "true",
		"enabledCategories": "TYPOS",
	})
	require.NoError(t, err)
	// grammar match should be filtered out when only TYPOS enabled
	require.NotContains(t, r.Body, "EN_A_VS_AN")
}

func TestFilterLocalsByCategories(t *testing.T) {
	ms := []languagetool.LocalMatch{
		{RuleID: "EN_A_VS_AN", Message: "a/an"},
		{RuleID: "MORFOLOGIK_RULE_EN_US", Message: "spell"},
	}
	out := filterLocalsByCategories(ms, CheckOptions{DisabledCategories: []string{"GRAMMAR"}})
	require.Len(t, out, 1)
	require.Equal(t, "MORFOLOGIK_RULE_EN_US", out[0].RuleID)

	out = filterLocalsByCategories(ms, CheckOptions{
		UseEnabledOnly:    true,
		EnabledCategories: []string{"GRAMMAR"},
	})
	require.Len(t, out, 1)
	require.Equal(t, "EN_A_VS_AN", out[0].RuleID)
}

func TestApiV2_MatchTypeAndContextForSure(t *testing.T) {
	api := NewApiV2(nil, nil)
	r, err := api.Handle("check", map[string]string{
		"language": "en",
		"text":     "This is an test.",
	})
	require.NoError(t, err)
	require.Contains(t, r.Body, `"typeName":"grammar"`)
	// contextForSureMatch omitted when 0 (omitempty); style path uses -1
	r2, err := api.Handle("check", map[string]string{
		"language":   "en",
		"text":       "word word word word word word word word.",
		"ruleValues": "TOO_LONG_SENTENCE:3",
	})
	require.NoError(t, err)
	require.Contains(t, r2.Body, "contextForSureMatch")
}
