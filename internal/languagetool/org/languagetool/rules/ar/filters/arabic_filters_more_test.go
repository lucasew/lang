package filters

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	ar_synth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ar"
	"github.com/stretchr/testify/require"
)

func TestMasdarToVerbAndMafoul(t *testing.T) {
	// Masdar→verb map is official file (exact keys); full Accept path covered by ArabicMasdarToVerbFilterTest_test.go.
	f := NewArabicMasdarToVerbFilter()
	require.Equal(t, []string{"عَمِلَ"}, f.Masdar2VerbList["عمل"])
	// Java authorizeLemma is exact "قَامَ" only
	require.Equal(t, []string{"قَامَ"}, f.filterLemmas([]string{"قَامَ", "ذهب", "قام"}))

	// Verb→masdar map is official file (exact keys); full Accept path covered by ArabicVerbToMafoulMutlaqFilterTest_test.go.
	v := NewArabicVerbToMafoulMutlaqFilter()
	// Last-wins for duplicate عَمِلَ lines in arabic_verb_masdar.txt
	require.Equal(t, []string{"عملة", "عمل"}, v.Verb2MasdarList["عَمِلَ"])
	require.Equal(t, []string{"إعمال"}, v.Verb2MasdarList["أَعْمَلَ"])
	// No tashkeel-stripped invent key
	require.Nil(t, v.Verb2MasdarList["عمل"])
}

func TestArabicMasdarToVerb_AcceptInjectedTags(t *testing.T) {
	// Injected readings with authorized lemma + masdar lemma → real synthesizer.InflectLemmaLike.
	f := NewArabicMasdarToVerbFilter()
	require.NotNil(t, f.synthesizer)
	// POS must pass tagger.GetLemmas filters: verb (IsVerb) / masdar (IsMasdar).
	// Use real tagger path instead of invent stub InflectLemmaLike.
	vpos := "V-1------"
	// NM is masdar class in ArabicTagManager.IsMasdar
	mpos := "NM------"
	auxLem, masdLem := "قَامَ", "عمل"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("يقوم", &vpos, &auxLem), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("بالعمل", &mpos, &masdLem), 4),
	}
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"verb": "يقوم", "noun": "بالعمل"}, 0, toks, nil)
	require.NotNil(t, out)
	// With real synth, form is inflected like the aux token (not bare lemma).
	require.NotEmpty(t, out.GetSuggestedReplacements())
}

func TestArabicMasdarToVerb_FailClosedWithoutTags(t *testing.T) {
	f := NewArabicMasdarToVerbFilter()
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
	// Official arabic_adjective_exclamation.txt via SimpleReplaceDataLoader (not invent طويل→أطول map).
	// Java loadWords does not trim; official lines are " جميل=…" (leading space on keys).
	require.NotEmpty(t, f.Adj2Comp)
	require.Equal(t, ArabicAdjectiveExclamationFileName, GetDataFilePath())
	require.Equal(t, []string{"ما أجمل", "أجمِل ب"}, f.ComparativesFor(" جميل"))
	require.Equal(t, []string{"ما أسعد", "أسعِد ب"}, f.ComparativesFor(" سعيد"))
	require.Equal(t, []string{"ما أشدّ استعجال", "اشدد باستعجال"}, f.ComparativesFor(" مستعجل"))
	// Exact get only — plain lemma (no leading space) misses, matching Java adj2compList.get
	require.Nil(t, f.ComparativesFor("جميل"))
	// No RemoveTashkeel invent fallback
	require.Nil(t, f.ComparativesFor("جَمِيل"))

	// prepareSuggestions matrix (Java prepareSuggestions(comp, noun))
	require.Equal(t, []string{"أطول"}, PrepareExclamationSuggestions("أطول", ""))
	require.Equal(t, []string{"أطول الولد"}, PrepareExclamationSuggestions("أطول", "الولد"))
	require.Equal(t, []string{"أطولني"}, PrepareExclamationSuggestions("أطول", "أنا")) // ArabicWordMaps: أنا→ني
	require.Equal(t, []string{"أطوله"}, PrepareExclamationSuggestions("أطول", "هو"))
	require.Equal(t, []string{"أطولها"}, PrepareExclamationSuggestions("أطول", "هي"))
	// Second form ends with " ب" → no extra space before non-pronoun noun
	require.Equal(t, []string{"أجمِل بالولد"}, PrepareExclamationSuggestions("أجمِل ب", "الولد"))
	// isPronoun is exact list only: هن/نحن are NOT pronouns here (dead private map only)
	require.Equal(t, []string{"أطول هن"}, PrepareExclamationSuggestions("أطول", "هن"))
	require.Equal(t, []string{"أطول نحن"}, PrepareExclamationSuggestions("أطول", "نحن"))
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
		"org.languagetool.rules.ar.filters.ArabicDMYDateCheckFilter",
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
	m.ShortMessage = "short"
	// Official SimpleReplace keys keep leading space (" جميل"); inject that lemma + NA POS.
	lem := " جميل"
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
	require.Equal(t, "msg", out.GetMessage())
	require.Equal(t, "short", out.ShortMessage)
	// prepareSuggestions: ما أجمله + أجمِل به (pronoun هو → ه attached, no space)
	require.ElementsMatch(t, []string{"ما أجمله", "أجمِل به"}, out.GetSuggestedReplacements())
}

func TestArabicAdjectiveExclamation_AcceptNonPronounNoun(t *testing.T) {
	f := NewArabicAdjectiveToExclamationFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	lem := " جميل"
	pos := "NA------"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("كم", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("الولد", nil, nil), 3),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("جميل", &pos, &lem), 9),
	}
	out := f.AcceptRuleMatch(m, map[string]string{
		"adj": "جميل", "adj_pos": "3", "noun": "الولد",
	}, 0, toks, nil)
	require.NotNil(t, out)
	// space after "ما أجمل"; no extra space after "أجمِل ب"
	require.ElementsMatch(t, []string{"ما أجمل الولد", "أجمِل بالولد"}, out.GetSuggestedReplacements())
}

func TestArabicAdjectiveExclamation_AcceptEmptyNoun(t *testing.T) {
	f := NewArabicAdjectiveToExclamationFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	lem := " جميل"
	pos := "NA------"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("جميل", &pos, &lem), 0),
	}
	out := f.AcceptRuleMatch(m, map[string]string{
		"adj": "جميل", "adj_pos": "1", "noun": "",
	}, 0, toks, nil)
	require.NotNil(t, out)
	require.ElementsMatch(t, []string{"ما أجمل", "أجمِل ب"}, out.GetSuggestedReplacements())
}

func TestArabicAdjectiveExclamation_FailClosedWithoutLemma(t *testing.T) {
	f := NewArabicAdjectiveToExclamationFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	// Surface only — Java getLemmas empty → no invent from surface / args adj.
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

func TestArabicAdjectiveExclamation_FailClosedPlainLemmaMissesSpacedKey(t *testing.T) {
	// Java exact get: lemma "جميل" does not hit key " جميل" from official file.
	f := NewArabicAdjectiveToExclamationFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	lem := "جميل"
	pos := "NA------"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("جميل", &pos, &lem), 0),
	}
	out := f.AcceptRuleMatch(m, map[string]string{
		"adj": "جميل", "adj_pos": "1", "noun": "هو",
	}, 0, toks, nil)
	require.NotNil(t, out)
	require.Empty(t, out.GetSuggestedReplacements())
}

func TestArabicAdjectiveExclamation_AdjPosParsePanic(t *testing.T) {
	f := NewArabicAdjectiveToExclamationFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("x", nil, nil), 0),
	}
	require.PanicsWithValue(t, "Error parsing adj_pos from : not-a-number", func() {
		f.AcceptRuleMatch(m, map[string]string{
			"adj": "x", "adj_pos": "not-a-number", "noun": "هو",
		}, 0, toks, nil)
	})
}

func TestArabicVerbToMafoul_Accept(t *testing.T) {
	f := NewArabicVerbToMafoulMutlaqFilter()
	require.NotNil(t, f.synthesizer)
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	lem := "عَمِلَ"
	// POS must pass tagger.GetLemmas IsVerb (starts with V).
	pos := "V-1------"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("يعمل", &pos, &lem), 0),
	}
	out := f.AcceptRuleMatch(m, map[string]string{"verb": "يعمل", "adj": "قوي"}, 0, toks, nil)
	require.NotNil(t, out)
	require.NotEmpty(t, out.GetSuggestedReplacements())
	// Java: verb + inflectMafoulMutlq(masdar) + inflectAdjectiveTanwinNasb(adj)
	// Last-wins map: عَمِلَ → عملة|عمل
	wantFem := "يعمل " + ar_synth.InflectMafoulMutlq("عملة") + " " + ar_synth.InflectAdjectiveTanwinNasb("قوي", true)
	wantMasc := "يعمل " + ar_synth.InflectMafoulMutlq("عمل") + " " + ar_synth.InflectAdjectiveTanwinNasb("قوي", false)
	require.Contains(t, out.GetSuggestedReplacements(), wantFem)
	require.Contains(t, out.GetSuggestedReplacements(), wantMasc)
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

func TestArabicVerbToMafoul_FailClosedPlainLemmaMissesTashkeelKey(t *testing.T) {
	// Java exact get: lemma without diacritics does not hit keyed "عَمِلَ".
	f := NewArabicVerbToMafoulMutlaqFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	lem := "عمل"
	pos := "V-1------"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("يعمل", &pos, &lem), 0),
	}
	out := f.AcceptRuleMatch(m, map[string]string{"verb": "يعمل", "adj": "قوي"}, 0, toks, nil)
	require.NotNil(t, out)
	require.Empty(t, out.GetSuggestedReplacements())
}
