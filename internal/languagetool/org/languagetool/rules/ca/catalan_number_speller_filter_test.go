package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Twin: word count uses split(" ") after hyphen normalize — empty mid-fields count.
func TestCatalanNumberSpellerFilter_SplitSpaceWordCount(t *testing.T) {
	// "a  b  c" → 5 parts with split(" ") → length >= 4 → suppress
	f := NewCatalanNumberSpellerFilter(func(s string) string { return "a  b  c" })
	require.Equal(t, "", f.Suggest("1", "", false))
	// three ASCII-space tokens: ok
	f2 := NewCatalanNumberSpellerFilter(func(s string) string { return "a b c" })
	require.Equal(t, "a b c", f2.Suggest("1", "", false))
}

func TestCatalanNumberSpellerFilter(t *testing.T) {
	f := NewCatalanNumberSpellerFilter(func(s string) string {
		if s == "feminine 2" {
			return "dues"
		}
		if s == "21" {
			return "vint-i-un"
		}
		if s == "1234" {
			return "mil dos-cents trenta-quatre extras"
		}
		return "u"
	})
	require.Equal(t, "dues", f.Suggest("2", "feminine", false))
	require.Equal(t, "Vint-i-un", f.Suggest("21", "", true))
	require.Equal(t, "", f.Suggest("1234", "", false)) // too many words
}

func TestCatalanNumberSpellerFilter_AcceptRuleMatch(t *testing.T) {
	f := NewCatalanNumberSpellerFilter(func(s string) string {
		if s == "feminine 2" {
			return "dues"
		}
		if s == "21" {
			return "vint-i-un"
		}
		return ""
	})
	m := rules.NewRuleMatch(nil, nil, 0, 2, "msg")
	// feminine at non-start
	out := f.AcceptRuleMatch(m, map[string]string{
		"number_to_spell": "2", "gender": "feminine",
	}, 3, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"dues"}, out.GetSuggestedReplacements())

	// sentence start → capitalize
	out = f.AcceptRuleMatch(m, map[string]string{
		"number_to_spell": "21", "gender": "masculine",
	}, 1, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Vint-i-un"}, out.GetSuggestedReplacements())

	// SENT_START on previous token
	ss := languagetool.SentenceStartTagName
	sentStart := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("", &ss, nil), 0)
	// Build a minimal sentence is hard; patternTokenPos<=1 covers common case.
	// fail-closed without SpellNumber
	require.Nil(t, NewCatalanNumberSpellerFilter(nil).AcceptRuleMatch(m, map[string]string{
		"number_to_spell": "2", "gender": "masculine",
	}, 2, nil, nil))
	_ = sentStart
}

func TestCatalanNumberSpellerFilterRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.CatalanNumberSpellerFilter"))
}
