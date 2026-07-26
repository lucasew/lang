package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianVerbConjugationRuleTest.java
// Java: new RussianVerbConjugationRule(messages) + JLanguageTool(Russian.getInstance()).getAnalyzedSentence
// Go: NewRussianVerbConjugationRule + NewJLanguageTool("ru") with RussianTagger wired (real morph).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	tagru "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ru"
	"github.com/stretchr/testify/require"
)

// Java rightSentences (expect 0 matches).
var russianVerbConjugationRightSentences = []string{
	"Я иду", "Она сидит", "Оно думает",
	"Они пишут", "Мы думаем", "Ты читаешь", "Он творит", "Вы идёте",
	"Я ходил", "Они ходили", "Мы ходили", "Она ходила", "Оно ходило", "Я ходила",
	"Я пойду", "Она пойдёт", "Оно пойдёт", "Мы пойдём", "Ты пойдёшь",
	"Я согласился на предложение.", "Джек и я согласились", "Ты может быть не помнишь.",
}

// Java wrongSentences (expect 1 match each).
var russianVerbConjugationWrongSentences = []string{
	"Я идёт", "Она сидят",
	"Оно думаешь", "Они идёте", "Мы думаю", "Ты читает", "Он творю",
	"Я ходили", "Они ходил", "Мы ходила", "Она ходил", "Оно ходила", "Я ходило",
	"Я пойдёт", "Она пойдут", "Оно пойдёте", "Мы пойдёшь", "Ты пойду",
	"Мы может поговорить здесь.",
}

// wireRUVerbConjugationLT ports JLanguageTool(Russian): tagger from RussianTagger.INSTANCE.
// Returns nil when russian.dict is unavailable (test skip).
func wireRUVerbConjugationLT(t *testing.T) *languagetool.JLanguageTool {
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

// Port of RussianVerbConjugationRuleTest.testRussianVerbConjugationRule
func TestRussianVerbConjugationRule_RussianVerbConjugationRule(t *testing.T) {
	rule := NewRussianVerbConjugationRule(nil)
	lt := wireRUVerbConjugationLT(t)

	for _, sentence := range russianVerbConjugationWrongSentences {
		analyzed := lt.GetAnalyzedSentence(sentence)
		require.NotNil(t, analyzed, "analyze %q", sentence)
		matches := rule.Match(analyzed)
		require.Equal(t, 1, len(matches), "Expected error in sentence: %s", sentence)
	}

	for _, sentence := range russianVerbConjugationRightSentences {
		analyzed := lt.GetAnalyzedSentence(sentence)
		require.NotNil(t, analyzed, "analyze %q", sentence)
		matches := rule.Match(analyzed)
		require.Equal(t, 0, len(matches), "Did not expect error in sentence: %s", sentence)
	}
}

// Metadata twin (Java constructor + getId/getDescription/getShort/category/example).
func TestRussianVerbConjugationRule_Metadata(t *testing.T) {
	r := NewRussianVerbConjugationRule(nil)
	require.Equal(t, "RU_VERB_CONJUGATION", r.GetID())
	require.Equal(t, "Согласование личных местоимений с глаголами", r.GetDescription())
	require.Equal(t, "Неверное спряжение глагола", r.GetShort())
	require.NotNil(t, r.GetCategory())
	require.True(t, r.GetCategory().GetID().Equals(rules.CategoryGrammar))
	require.NotEmpty(t, r.GetIncorrectExamples())
	require.NotEmpty(t, r.GetCorrectExamples())
	require.Equal(t, []string{"Я иду"}, r.GetIncorrectExamples()[0].GetCorrections())
	// Match message on a known wrong sentence.
	lt := wireRUVerbConjugationLT(t)
	ms := r.Match(lt.GetAnalyzedSentence("Я идёт"))
	require.Len(t, ms, 1)
	require.Equal(t, "Неверное спряжение глагола или неверное местоимение", ms[0].GetMessage())
	require.Equal(t, "Неверное спряжение глагола", ms[0].GetShortMessage())
}
