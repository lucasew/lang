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
	// Java authorizeLemma is exact "قَامَ" only
	require.Equal(t, []string{"قَامَ"}, FilterAuxLemmas([]string{"قَامَ", "ذهب", "قام"}))

	m := rules.NewRuleMatch("r", nil, 0, 1, "msg")
	ApplySuggestions(m, []string{"أ", "ب"})
	require.Equal(t, []string{"أ", "ب"}, m.GetSuggestedReplacements())

	v := NewArabicVerbToMafoulMutlaqFilter()
	sug := v.SuggestMafoulMutlaq("عَمِلَ")
	// Java inflectMafoulMutlq: عمل + fathatan + alef
	require.Contains(t, sug, ar_synth.InflectMafoulMutlq("عمل"))
}

func TestArabicMasdarToVerb_Accept(t *testing.T) {
	f := NewArabicMasdarToVerbFilter()
	// Without InflectLemmaLike: fail closed
	vpos, mpos := "V", "NM------"
	auxLem, masdLem := "قَامَ", "عمل"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("قام", &vpos, &auxLem), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("بالأكل", &mpos, &masdLem), 4),
	}
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"verb": "قام", "noun": "أكل"}, 0, toks, nil)
	require.NotNil(t, out)
	require.Empty(t, out.GetSuggestedReplacements())

	// With InflectLemmaLike stub (stand-in for ArabicSynthesizer)
	f.InflectLemmaLike = func(targetLemma string, source *languagetool.AnalyzedToken) []string {
		return []string{targetLemma}
	}
	out = f.AcceptRuleMatch(m, map[string]string{"verb": "قام", "noun": "أكل"}, 0, toks, nil)
	require.Contains(t, out.GetSuggestedReplacements(), "عَمِلَ")
}

func TestArabicMasdarToVerb_FailClosedWithoutTags(t *testing.T) {
	f := NewArabicMasdarToVerbFilter()
	f.InflectLemmaLike = func(targetLemma string, source *languagetool.AnalyzedToken) []string {
		return []string{targetLemma}
	}
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("قام", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("بالأكل", nil, nil), 4),
	}
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"verb": "قام", "noun": "أكل"}, 0, toks, nil)
	require.NotNil(t, out)
	require.Empty(t, out.GetSuggestedReplacements())
}

func TestAdjectiveExclamation(t *testing.T) {
	f := NewArabicAdjectiveToExclamationFilter()
	// Official arabic_adjective_exclamation.txt (not invent طويل→أطول map)
	require.NotEmpty(t, f.Adj2Comp)
	require.Equal(t, []string{"ما أجمل", "أجمِل ب"}, f.ComparativesFor("جميل"))
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
	// Official map: جميل=ما أجمل|أجمِل ب — inject adj lemma/POS (no surface invent).
	lem := "جميل"
	pos := "NA------"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("كم", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("هو", nil, nil), 3),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("جميل", &pos, &lem), 6),
	}
	out := f.AcceptRuleMatch(m, map[string]string{
		"adj": "جميل", "adj_pos": "3", "noun": "هو",
	}, 0, toks, nil)
	require.NotNil(t, out)
	require.NotEmpty(t, out.GetSuggestedReplacements())
	// PrepareExclamationSuggestions attaches pronoun suffix to comparative forms
	require.True(t, len(out.GetSuggestedReplacements()) >= 1)
}

func TestArabicAdjectiveExclamation_FailClosedWithoutLemma(t *testing.T) {
	f := NewArabicAdjectiveToExclamationFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("كم", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("هو", nil, nil), 3),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("جميل", nil, nil), 6),
	}
	out := f.AcceptRuleMatch(m, map[string]string{
		"adj": "جميل", "adj_pos": "3", "noun": "هو",
	}, 0, toks, nil)
	require.NotNil(t, out)
	require.Empty(t, out.GetSuggestedReplacements())
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
