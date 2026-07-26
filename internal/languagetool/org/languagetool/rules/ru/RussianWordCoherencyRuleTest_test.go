package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianWordCoherencyRuleTest.java
//
// Java setup:
//   private final JLanguageTool lt = new JLanguageTool(Russian.getInstance());
//   assertGood/assertError → new RussianWordCoherencyRule(messages).match(singletonList(lt.getAnalyzedSentence(s)))
//   testRuleCompleteTexts → lt.check(text).size() with only coherency differences in those fixtures
//
// Go: NewJLanguageTool("ru") + RussianTagger (real russian.dict lemmas).
// Complete-text cases run RussianWordCoherencyRule.MatchList(lt.Analyze(text)) — same text-level
// span as Java AbstractWordCoherencyRule.match over analyzed sentences (EN WordCoherency twin pattern).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	tagru "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ru"
	"github.com/stretchr/testify/require"
)

// wireRUWordCoherencyLT ports JLanguageTool(Russian.getInstance()) for this rule twin.
// Returns nil when russian.dict is unavailable (test skip).
func wireRUWordCoherencyLT(t *testing.T) *languagetool.JLanguageTool {
	t.Helper()
	if tagru.DiscoverRussianPOSDict() == "" {
		t.Skip("russian.dict not in tree")
	}
	tagru.EnsureDefaultRussianTagger()
	require.NotNil(t, tagru.DefaultRussianTagger)
	require.NotNil(t, tagru.DefaultRussianTagger.GetWordTagger())
	require.NotEmpty(t, tagru.RussianPOSDictPath(), "real russian.dict must load")

	lt := languagetool.NewJLanguageTool("ru")
	// Java: language.getTagger().tag(tokens) — full-sentence RussianTagger.Tag path.
	lt.SentenceTagger = tagru.DefaultRussianTagger.Tag
	return lt
}

// TestRussianWordCoherencyRule_Rule twins Java RussianWordCoherencyRuleTest.testRule.
func TestRussianWordCoherencyRule_Rule(t *testing.T) {
	lt := wireRUWordCoherencyLT(t)

	assertGood := func(s string) {
		t.Helper()
		rule := NewRussianWordCoherencyRule(nil)
		analyzed := lt.GetAnalyzedSentence(s)
		require.NotNil(t, analyzed, "analyze %q", s)
		require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{analyzed})), "good %q", s)
	}
	assertError := func(s string) {
		t.Helper()
		rule := NewRussianWordCoherencyRule(nil)
		analyzed := lt.GetAnalyzedSentence(s)
		require.NotNil(t, analyzed, "analyze %q", s)
		require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{analyzed})), "error %q", s)
	}

	assertGood("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C.")
	assertGood("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C.")
	assertError("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C или ноль по шкале Кельвина.")
}

// TestRussianWordCoherencyRule_CallIndependence twins Java testCallIndependence.
func TestRussianWordCoherencyRule_CallIndependence(t *testing.T) {
	lt := wireRUWordCoherencyLT(t)
	assertGood := func(s string) {
		t.Helper()
		rule := NewRussianWordCoherencyRule(nil)
		analyzed := lt.GetAnalyzedSentence(s)
		require.NotNil(t, analyzed, "analyze %q", s)
		require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{analyzed})))
	}
	assertGood("Абсолютный нуль.")
	assertGood("Ноль по шкале Кельвина.") // this won't be noticed, the calls are independent of each other
}

// TestRussianWordCoherencyRule_RuleCompleteTexts twins Java testRuleCompleteTexts.
// Java: assertEquals(N, lt.check(text).size()) with RU_WORD_COHERENCY as the only hit on these texts.
// Go: MatchList over lt.Analyze (document-relative text-level match; same N as Java).
func TestRussianWordCoherencyRule_RuleCompleteTexts(t *testing.T) {
	lt := wireRUWordCoherencyLT(t)
	check := func(s string) int {
		t.Helper()
		return len(NewRussianWordCoherencyRule(nil).MatchList(lt.Analyze(s)))
	}
	require.Equal(t, 0, check("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C или нуль по шкале Кельвина."))
	require.Equal(t, 1, check("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C или ноль по шкале Кельвина."))
	// cross-paragraph checks:
	require.Equal(t, 1, check("Абсолютный нуль.\n\nСовсем недостижим. И ноль по шкале Кельвина."))
}

// TestRussianWordCoherencyRule_Metadata twins Java id/description/message/example pair.
func TestRussianWordCoherencyRule_Metadata(t *testing.T) {
	rule := NewRussianWordCoherencyRule(nil)
	require.Equal(t, "RU_WORD_COHERENCY", rule.GetID())
	require.Equal(t, "Единообразное написание слов с более чем одним допустимым написанием", rule.GetDescription())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Понятие «оффлайн» тоже имеет английские корни и связано со словом «offline», что означает «вне сети». Принтер перешёл в состояние <marker>офлайн</marker>.", inc[0].GetExample())
	require.Equal(t, []string{"оффлайн"}, inc[0].GetCorrections())
	require.Equal(t, "Понятие «оффлайн» тоже имеет английские корни и связано со словом «offline», что означает «вне сети». Принтер перешёл в состояние <marker>оффлайн</marker>.", rule.GetCorrectExamples()[0].GetExample())

	lt := wireRUWordCoherencyLT(t)
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{
		lt.GetAnalyzedSentence("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C или ноль по шкале Кельвина."),
	})
	require.Len(t, ms, 1)
	require.Equal(t, "«ноль» и «нуль» не следует использовать одновременно", ms[0].GetMessage())
}
