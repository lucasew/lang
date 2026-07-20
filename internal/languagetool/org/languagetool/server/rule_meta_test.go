package server

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

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
	// Java AvsAnRule → Categories.MISC; Morfologik speller → TYPOS.
	ms := []languagetool.LocalMatch{
		{RuleID: "EN_A_VS_AN", Message: "a/an"},
		{RuleID: "MORFOLOGIK_RULE_EN_US", Message: "spell"},
	}
	out := filterLocalsByCategories(ms, CheckOptions{DisabledCategories: []string{"MISC"}})
	require.Len(t, out, 1)
	require.Equal(t, "MORFOLOGIK_RULE_EN_US", out[0].RuleID)

	out = filterLocalsByCategories(ms, CheckOptions{
		UseEnabledOnly:    true,
		EnabledCategories: []string{"MISC"},
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
	// Java AvsAnRule: ITSIssueType.Misspelling
	require.Contains(t, r.Body, `"typeName":"misspelling"`)
	// contextForSureMatch omitted when 0 (omitempty); text-level rules (Java -1) serialize it.
	// LongSentenceRule is Tag.picky — enable Level.PICKY so the match is kept.
	r2, err := api.Handle("check", map[string]string{
		"language":   "en",
		"level":      "picky",
		"text":       "word word word word word word word word.",
		"ruleValues": "TOO_LONG_SENTENCE:3",
	})
	require.NoError(t, err)
	require.Contains(t, r2.Body, "contextForSureMatch")
	require.Contains(t, r2.Body, `"contextForSureMatch":-1`)
}

func TestApiV2_MatchRuleURL(t *testing.T) {
	api := NewApiV2(nil, nil)
	r, err := api.Handle("check", map[string]string{
		"language": "en",
		"text":     "This is an test.",
	})
	require.NoError(t, err)
	require.Contains(t, r.Body, "community.languagetool.org/rule/show/EN_A_VS_AN")
	require.Contains(t, r.Body, "lang=en")
}

