package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicWordCoherencyRuleTest.java
//
// Java setup:
//   private final JLanguageTool lt = new JLanguageTool(Languages.getLanguageForShortCode("ar"));
//   @Before: TestTools.disableAllRulesExcept(lt, "AR_WORD_COHERENCY");
//   testRule body has only a commented-out assertError — still provide outcome-fidelity
//   twins from Java-visible behavior (id, description, category, example pair, official
//   data pairs, MatchList counts for consistent vs mixed spelling).
//
// Go analysis path (same morph lemmas Java uses):
//   NewJLanguageTool("ar") + ArabicWordTokenizer (via WordTokenizerForLanguage) +
//   ArabicTagger.Tag as SentenceTagger (arabic.dict lemmas).
// Production loads /ar/coherency.txt pairs only (no invent expand).
// Java example works because tagger lemmas الشؤون → شؤون, then later surface شئون conflicts.
// Match outcomes only assert pairs that fire under real AR lemmas (no invent lemmas for
// forms the tagger rewrites away from the file keys, e.g. الذي → لَذِيَ, تركية → ترك).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	tagar "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
	"github.com/stretchr/testify/require"
)

// wireARWordCoherencyLT ports JLanguageTool(Languages.getLanguageForShortCode("ar"))
// for this rule twin. Returns nil when arabic.dict is unavailable (test skip).
func wireARWordCoherencyLT(t *testing.T) *languagetool.JLanguageTool {
	t.Helper()
	if tagar.DiscoverArabicPOSDict() == "" {
		t.Skip("arabic.dict not in tree")
	}
	tagar.EnsureDefaultArabicTagger()
	require.NotNil(t, tagar.DefaultArabicTagger)
	require.NotNil(t, tagar.DefaultArabicTagger.GetWordTagger())

	lt := languagetool.NewJLanguageTool("ar")
	// Java: language.getTagger().tag(tokens) — full-sentence ArabicTagger.Tag path.
	lt.SentenceTagger = tagar.DefaultArabicTagger.Tag
	return lt
}

// analyzeARWordCoherencySentence ports lt.getAnalyzedSentence for Arabic.
func analyzeARWordCoherencySentence(t *testing.T, lt *languagetool.JLanguageTool, text string) *languagetool.AnalyzedSentence {
	t.Helper()
	s := lt.GetAnalyzedSentence(text)
	require.NotNil(t, s, "analyze %q", text)
	return s
}

// TestArabicWordCoherencyRule_Rule twins Java testRule outcomes from official pairs
// and the Java example pair (assertError body is commented-out upstream).
func TestArabicWordCoherencyRule_Rule(t *testing.T) {
	lt := wireARWordCoherencyLT(t)
	rule := NewArabicWordCoherencyRule(nil)

	assertGood := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
			analyzeARWordCoherencySentence(t, lt, s),
		})), "good: %q", s)
	}
	assertError := func(s string) {
		t.Helper()
		require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{
			analyzeARWordCoherencySentence(t, lt, s),
		})), "error: %q", s)
	}

	// consistent spelling
	assertGood("كل شؤون العالم.")
	assertGood("كل شئون الناس.")
	assertGood("وزارة الشؤون الخارجية تهتم  بكل شؤون العالم.")
	assertGood("مسؤول")
	assertGood("مسئول")
	// Java example pair wrong (without markers): الشؤون (lemma شؤون) then شئون
	assertError("وزارة الشؤون الخارجية تهتم  بكل شئون العالم.")
	// official pair شؤون;شئون as surface/lemma forms
	assertError("كل شؤون العالم وكل شئون الناس.")
	// reverse order
	assertError("كل شئون الناس وكل شؤون العالم.")
	// other official pairs that keep file keys under ArabicTagger lemmas
	assertError("مسؤول مسئول")
	assertError("مسئول مسؤول")
	assertError("ظَنَّ ضَنَّ")
}

// TestArabicWordCoherencyRule_Metadata twins Java id/description/MISC/example pair.
func TestArabicWordCoherencyRule_Metadata(t *testing.T) {
	rule := NewArabicWordCoherencyRule(nil)
	require.Equal(t, "AR_WORD_COHERENCY", rule.GetID())
	require.Equal(t, "ضبط انسجام التهجئة للكلمات التي تكتب بطرق مختلفة مقبولة.", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.NewCategoryId("MISC"), rule.GetCategory().GetID())
	require.Equal(t, -1, rule.MinToCheckParagraph())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "وزارة الشؤون الخارجية تهتم  بكل <marker>شئون</marker> العالم.", inc[0].GetExample())
	require.Equal(t, []string{"شؤون"}, inc[0].GetCorrections())
	require.Equal(t,
		"وزارة الشؤون الخارجية تهتم  بكل <marker>شؤون</marker> العالم.",
		rule.GetCorrectExamples()[0].GetExample(),
	)
}

// TestArabicWordCoherencyRule_DataFilePairs asserts official /ar/coherency.txt bidirectional map.
func TestArabicWordCoherencyRule_DataFilePairs(t *testing.T) {
	d := loadCoherencyData()
	// Java example pair basis
	require.Contains(t, d.WordMap, "شؤون")
	require.Contains(t, d.WordMap["شؤون"], "شئون")
	require.Contains(t, d.WordMap, "شئون")
	require.Contains(t, d.WordMap["شئون"], "شؤون")
	// other official pairs (resource parity; loader is bidirectional)
	require.Contains(t, d.WordMap["الشؤون"], "الشئون")
	require.Contains(t, d.WordMap["الشئون"], "الشؤون")
	require.Contains(t, d.WordMap["ظَنَّ"], "ضَنَّ")
	require.Contains(t, d.WordMap["ضَنَّ"], "ظَنَّ")
	require.Contains(t, d.WordMap["مسؤول"], "مسئول")
	require.Contains(t, d.WordMap["مسئول"], "مسؤول")
	require.Contains(t, d.WordMap["الذي"], "الذى")
	require.Contains(t, d.WordMap["الذى"], "الذي")
	require.Contains(t, d.WordMap["تركية"], "تركيا")
	require.Contains(t, d.WordMap["تركيا"], "تركية")
}

// TestArabicWordCoherencyRule_MessageAndSuggestion twins Java getMessage + suggestion.
func TestArabicWordCoherencyRule_MessageAndSuggestion(t *testing.T) {
	lt := wireARWordCoherencyLT(t)
	rule := NewArabicWordCoherencyRule(nil)

	// Java example: established الشؤون→lemma شؤون → later شئون
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzeARWordCoherencySentence(t, lt, "وزارة الشؤون الخارجية تهتم  بكل شئون العالم."),
	})
	require.Len(t, ms, 1)
	// Java message: double space after "في"
	require.Equal(t, "تجنب استعمال شكلين للكلمة نفسها ('شئون' و 'شؤون') في  النص نفسه.", ms[0].GetMessage())
	require.Equal(t, 32, ms[0].GetFromPos())
	require.Equal(t, 36, ms[0].GetToPos())
	require.Equal(t, []string{"شؤون"}, ms[0].GetSuggestedReplacements())

	// surface pair: established شؤون → later شئون
	ms = rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzeARWordCoherencySentence(t, lt, "كل شؤون العالم وكل شئون الناس."),
	})
	require.Len(t, ms, 1)
	require.Equal(t, "تجنب استعمال شكلين للكلمة نفسها ('شئون' و 'شؤون') في  النص نفسه.", ms[0].GetMessage())
	require.Equal(t, []string{"شؤون"}, ms[0].GetSuggestedReplacements())

	// reverse: established شئون → later شؤون
	ms = rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzeARWordCoherencySentence(t, lt, "كل شئون الناس وكل شؤون العالم."),
	})
	require.Len(t, ms, 1)
	require.Equal(t, "تجنب استعمال شكلين للكلمة نفسها ('شؤون' و 'شئون') في  النص نفسه.", ms[0].GetMessage())
	require.Equal(t, []string{"شئون"}, ms[0].GetSuggestedReplacements())
}

// TestArabicWordCoherencyRule_CallIndependence: separate match calls do not share state.
func TestArabicWordCoherencyRule_CallIndependence(t *testing.T) {
	lt := wireARWordCoherencyLT(t)
	assertGood := func(s string) {
		t.Helper()
		rule := NewArabicWordCoherencyRule(nil)
		require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
			analyzeARWordCoherencySentence(t, lt, s),
		})))
	}
	assertGood("كل شؤون العالم.")
	assertGood("كل شئون الناس.") // not noticed across independent calls
}

// TestArabicWordCoherencyRule_RuleCompleteTexts: document-level MatchList over analyzeText.
func TestArabicWordCoherencyRule_RuleCompleteTexts(t *testing.T) {
	lt := wireARWordCoherencyLT(t)
	check := func(s string) int {
		t.Helper()
		return len(NewArabicWordCoherencyRule(nil).MatchList(lt.Analyze(s)))
	}
	require.Equal(t, 0, check("وزارة الشؤون الخارجية تهتم  بكل شؤون العالم."))
	require.Equal(t, 1, check("وزارة الشؤون الخارجية تهتم  بكل شئون العالم."))
	require.Equal(t, 1, check("كل شؤون العالم. كل شئون الناس."))
}
