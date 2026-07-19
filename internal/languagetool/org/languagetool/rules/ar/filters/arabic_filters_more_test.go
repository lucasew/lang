package filters

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	ar_synth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ar"
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
	// Java inflectMafoulMutlq: عمل + fathatan + alef
	require.Contains(t, sug, ar_synth.InflectMafoulMutlq("عمل"))
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

func TestArabicFiltersRegistered(t *testing.T) {
	for _, class := range []string{
		"org.languagetool.rules.ar.filters.ArabicDateCheckFilter",
		"org.languagetool.rules.ar.filters.ArabicAdvancedSynthesizerFilter",
		"org.languagetool.rules.ar.filters.ArabicNumberPhraseFilter",
		"org.languagetool.rules.ar.filters.ArabicMasdarToVerbFilter",
		"org.languagetool.rules.ar.filters.ArabicAdjectiveToExclamationFilter",
		"org.languagetool.rules.ar.filters.ArabicVerbToMafoulMutlaqFilter",
	} {
		require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	}
}

func TestArabicNumberPhraseFilter_AcceptRuleMatch(t *testing.T) {
	f := NewArabicNumberPhraseFilter()
	// previous + number digits, unit at end (nextPos -1)
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("في", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("3", nil, nil), 3),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("صندوق", nil, nil), 5),
	}
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"previous": "في", "previousPos": "1", "nextPos": "-1",
	}, 0, toks, nil)
	require.NotNil(t, out)
	require.NotEmpty(t, out.GetSuggestedReplacements())
	require.Contains(t, out.GetSuggestedReplacements()[0], "في")
}

func TestArabicAdjectiveExclamation_Accept(t *testing.T) {
	f := NewArabicAdjectiveToExclamationFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("كم", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("هو", nil, nil), 3),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("طويل", nil, nil), 6),
	}
	out := f.AcceptRuleMatch(m, map[string]string{
		"adj": "طويل", "adj_pos": "3", "noun": "هو",
	}, 0, toks, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "أطول"+"ه") // أطول + ه
}

func TestArabicVerbToMafoul_Accept(t *testing.T) {
	f := NewArabicVerbToMafoulMutlaqFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	lem := "عَمِلَ"
	pos := "V"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("يعمل", &pos, &lem), 0),
	}
	out := f.AcceptRuleMatch(m, map[string]string{"verb": "يعمل", "adj": "قوي"}, 0, toks, nil)
	require.NotNil(t, out)
	require.NotEmpty(t, out.GetSuggestedReplacements())
	// Java: verb + inflectMafoulMutlq(masdar) + inflectAdjectiveTanwinNasb(adj)
	want := "يعمل " + ar_synth.InflectMafoulMutlq("عمل") + " " + ar_synth.InflectAdjectiveTanwinNasb("قوي", false)
	require.Contains(t, out.GetSuggestedReplacements(), want)
}

func TestArabicVerbToMafoul_FailClosedWithoutLemma(t *testing.T) {
	f := NewArabicVerbToMafoulMutlaqFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	// Surface token only — Java tagger.getLemmas empty → no invent from surface.
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("يعمل", nil, nil), 0),
	}
	out := f.AcceptRuleMatch(m, map[string]string{"verb": "يعمل", "adj": "قوي"}, 0, toks, nil)
	require.NotNil(t, out)
	require.Empty(t, out.GetSuggestedReplacements())
}
