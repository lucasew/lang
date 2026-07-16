package filters

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestArabicDMYParse(t *testing.T) {
	d, m, y, err := ParseDMYDateArg("15-03-2024")
	require.NoError(t, err)
	require.Equal(t, "15", d)
	require.Equal(t, "03", m)
	require.Equal(t, "2024", y)
	_, _, _, err = ParseDMYDateArg("bad")
	require.Error(t, err)

	f := NewArabicDMYDateCheckFilter()
	_, err = f.AcceptDMYRuleMatch(nil, map[string]string{"date": "01-01-2024", "day": "1"})
	require.Error(t, err)
}

func TestMasdarToVerbAndMafoul(t *testing.T) {
	f := NewArabicMasdarToVerbFilter()
	require.Equal(t, []string{"عَمِلَ"}, f.SuggestVerbsForMasdar("عمل"))
	require.Contains(t, f.SuggestionsFromArgs(map[string]string{"noun": "عمل"}), "عَمِلَ")
	require.Equal(t, []string{"قَامَ"}, FilterAuxLemmas([]string{"قَامَ", "ذهب"}))

	m := rules.NewRuleMatch("r", nil, 0, 1, "msg")
	ApplySuggestions(m, []string{"أ", "ب"})
	require.Equal(t, []string{"أ", "ب"}, m.GetSuggestedReplacements())

	v := NewArabicVerbToMafoulMutlaqFilter()
	sug := v.SuggestMafoulMutlaq("عَمِلَ")
	require.Contains(t, sug, "عمل")
}

func TestAdjectiveExclamation(t *testing.T) {
	f := NewArabicAdjectiveToExclamationFilter()
	require.Equal(t, []string{"أطول"}, f.ComparativesFor("طويل"))
	require.Equal(t, []string{"أطول الولد"}, PrepareExclamationSuggestions("أطول", "الولد"))
	require.Equal(t, []string{"أطولني"}, PrepareExclamationSuggestions("أطول", "أنا"))
}

func TestAdvancedSynthFilterConstruct(t *testing.T) {
	f := NewArabicAdvancedSynthesizerFilter(func(lemma, postag string) []string {
		return []string{lemma + ":" + postag}
	})
	require.NotNil(t, f.AbstractAdvancedSynthesizerFilter)
}
