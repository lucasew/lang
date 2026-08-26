package ru

// Behavior twin for org.languagetool.rules.ru.RussianWordRootRepeatRule
// (no dedicated Java @Test — matrix from Java-visible ctor/match behavior).
//
// Java:
//   wordMap = WordCoherencyDataLoader.loadWords("/ru/wordrootrep.txt")
//   setDefaultOff(); getId = RU_WORD_ROOT_REPEAT; description = "Повтор однокоренных слов"
//   getMessage(w1,w2) = "«"+w1+"» и «"+w2+"» – однокоренные слова, их не стоит использовать одновременно"
//   example: абрикосный → абрикосов (style rewrite; MatchList suggestion uses createReplacement → "абрикос")
//
// Go: NewJLanguageTool("ru") + RussianTagger (real russian.dict lemmas), same pattern as
// RussianWordCoherencyRule twin. MatchList is invoked directly (default-off only affects
// JLanguageTool enablement at registration, not MatchList on the rule object).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	tagru "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ru"
	"github.com/stretchr/testify/require"
)

// wireRUWordRootRepeatLT ports JLanguageTool(Russian.getInstance()) for this rule twin.
func wireRUWordRootRepeatLT(t *testing.T) *languagetool.JLanguageTool {
	t.Helper()
	if tagru.DiscoverRussianPOSDict() == "" {
		t.Skip("russian.dict not in tree")
	}
	tagru.EnsureDefaultRussianTagger()
	require.NotNil(t, tagru.DefaultRussianTagger)
	require.NotNil(t, tagru.DefaultRussianTagger.GetWordTagger())
	require.NotEmpty(t, tagru.RussianPOSDictPath(), "real russian.dict must load")

	lt := languagetool.NewJLanguageTool("ru")
	lt.SentenceTagger = tagru.DefaultRussianTagger.Tag
	return lt
}

// TestRussianWordRootRepeatRule_Metadata twins Java id/description/defaultOff/example/message.
func TestRussianWordRootRepeatRule_Metadata(t *testing.T) {
	rule := NewRussianWordRootRepeatRule(nil)
	require.Equal(t, "RU_WORD_ROOT_REPEAT", rule.GetID())
	require.Equal(t, "Повтор однокоренных слов", rule.GetDescription())
	require.True(t, rule.IsDefaultOff(), "Java setDefaultOff()")

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Абрикос рос в саду. У меня на столе стоит <marker>абрикосный</marker> сок.", inc[0].GetExample())
	require.Equal(t, []string{"абрикосов"}, inc[0].GetCorrections())
	require.Equal(t, "Абрикос рос в саду. У меня на столе стоит сок из <marker>абрикосов</marker>.", rule.GetCorrectExamples()[0].GetExample())

	lt := wireRUWordRootRepeatLT(t)
	// Cross-sentence coherency on official pair абрикос;абрикосный (wordrootrep.txt).
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{
		lt.GetAnalyzedSentence("Абрикос рос в саду."),
		lt.GetAnalyzedSentence("У меня на столе стоит абрикосный сок."),
	})
	require.Len(t, ms, 1)
	require.Equal(t, "«абрикосный» и «абрикос» – однокоренные слова, их не стоит использовать одновременно", ms[0].GetMessage())
}

// TestRussianWordRootRepeatRule_CrossSentenceMatch twins Java example pair match behavior.
// default-off does not suppress MatchList on the rule object (same as coherency twin).
func TestRussianWordRootRepeatRule_CrossSentenceMatch(t *testing.T) {
	lt := wireRUWordRootRepeatLT(t)
	rule := NewRussianWordRootRepeatRule(nil)
	require.True(t, rule.IsDefaultOff())

	ms := rule.MatchList([]*languagetool.AnalyzedSentence{
		lt.GetAnalyzedSentence("Абрикос рос в саду."),
		lt.GetAnalyzedSentence("У меня на столе стоит абрикосный сок."),
	})
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetMessage(), "однокоренные")

	// Full-document Analyze path (two sentences).
	require.Equal(t, 1, len(rule.MatchList(lt.Analyze("Абрикос рос в саду. У меня на столе стоит абрикосный сок."))))
}

// TestRussianWordRootRepeatRule_NegativeNoConflict: single sentence without conflicting root pair → 0.
func TestRussianWordRootRepeatRule_NegativeNoConflict(t *testing.T) {
	lt := wireRUWordRootRepeatLT(t)
	rule := NewRussianWordRootRepeatRule(nil)

	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		lt.GetAnalyzedSentence("Абрикос рос в саду."),
	})))
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		lt.GetAnalyzedSentence("У меня на столе стоит абрикосный сок."),
	})))
	// Same root form twice is not a conflicting pair (map tracks the other spelling only).
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		lt.GetAnalyzedSentence("Абрикос и ещё один абрикос."),
	})))
}

// TestRussianWordRootRepeatRule_WordMapLoaded: official pair from wordrootrep.txt is present.
func TestRussianWordRootRepeatRule_WordMapLoaded(t *testing.T) {
	rule := NewRussianWordRootRepeatRule(nil)
	require.NotNil(t, rule.WordMap)
	alts, ok := rule.WordMap["абрикос"]
	require.True(t, ok, "абрикос must be in wordrootrep.txt map")
	_, has := alts["абрикосный"]
	require.True(t, has)
	alts2, ok2 := rule.WordMap["абрикосный"]
	require.True(t, ok2)
	_, has2 := alts2["абрикос"]
	require.True(t, has2)
}


// TestRussianWordRootRepeatRule_RegisterDefaultOff: Java setDefaultOff → MarkDefaultOff at register.
func TestRussianWordRootRepeatRule_RegisterDefaultOff(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ru")
	RegisterCoreRussianRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "RU_WORD_ROOT_REPEAT")
	require.Contains(t, lt.GetDefaultOffRuleIDs(), "RU_WORD_ROOT_REPEAT")
}
