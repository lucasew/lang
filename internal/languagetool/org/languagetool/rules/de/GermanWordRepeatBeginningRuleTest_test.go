package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanWordRepeatBeginningRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func deWRBMessages() map[string]string {
	return map[string]string{
		"desc_repetition_beginning_adv":       "Drei aufeinanderfolgende Sätze beginnen mit demselben Adverb.",
		"desc_repetition_beginning_word":      "Drei aufeinanderfolgende Sätze beginnen mit demselben Wort.",
		"desc_repetition_beginning_thesaurus": "Verwenden Sie ggf. einen Thesaurus.",
	}
}

func TestGermanWordRepeatBeginningRule_Rule(t *testing.T) {
	rule := NewGermanWordRepeatBeginningRule(deWRBMessages())

	// Java correct:
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Er ist nett. Er heißt Max."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Außerdem kommt er. Ferner kommt sie. Außerdem kommt es."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("2011: Dieses passiert. 2011: Jenes passiert. 2011: Nicht passiert"))))
	// Java errors:
	require.Equal(t, 1, len(rule.MatchList(languagetool.SplitAndAnalyze("Er ist nett. Er heißt Max. Er ist 11."))))
	require.Equal(t, 1, len(rule.MatchList(languagetool.SplitAndAnalyze("Außerdem kommt er. Außerdem kommt sie."))))
	// reset / single:
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Außerdem ist das ein neuer Text."))))
	// only real sentences ending [.?!]:
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Außerdem ist das ein neuer Text\n\nAußerdem noch mehr ohne Punkt\n\nAußerdem schon wieder"))))
}

func TestGermanWordRepeatBeginningRule_AdverbList(t *testing.T) {
	r := NewGermanWordRepeatBeginningRule(deWRBMessages())
	// Java ADVERBS exact surface (incl. Ü)
	require.True(t, r.isAdverb(atrWithPOS("Außerdem", "ADV", "außerdem")))
	require.True(t, r.isAdverb(atrWithPOS("Danach", "ADV", "danach")))
	require.True(t, r.isAdverb(atrWithPOS("Überdies", "ADV", "überdies")))
	require.False(t, r.isAdverb(atrWithPOS("Dann", "ADV", "dann")))
	require.False(t, r.isAdverb(nil))
}

func TestGermanWordRepeatBeginningRule_WordVsAdverbShortMsg(t *testing.T) {
	rule := NewGermanWordRepeatBeginningRule(deWRBMessages())

	// three "Er" → word short message
	ms := rule.MatchList(languagetool.SplitAndAnalyze("Er ist nett. Er heißt Max. Er ist 11."))
	require.Equal(t, 1, len(ms))
	require.Equal(t, deWRBMessages()["desc_repetition_beginning_word"], ms[0].GetShortMessage())
	require.Contains(t, ms[0].GetMessage(), deWRBMessages()["desc_repetition_beginning_thesaurus"])
	// match length UTF-16 of "Er"
	require.Equal(t, 2, ms[0].GetToPos()-ms[0].GetFromPos())

	// two "Außerdem" → adverb short message
	ms = rule.MatchList(languagetool.SplitAndAnalyze("Außerdem kommt er. Außerdem kommt sie."))
	require.Equal(t, 1, len(ms))
	require.Equal(t, deWRBMessages()["desc_repetition_beginning_adv"], ms[0].GetShortMessage())
	// "Außerdem" UTF-16 length (all BMP)
	require.Equal(t, len([]rune("Außerdem")), ms[0].GetToPos()-ms[0].GetFromPos())
}

func TestGermanWordRepeatBeginningRule_Category(t *testing.T) {
	r := NewGermanWordRepeatBeginningRule(deWRBMessages())
	require.NotNil(t, r.GetCategory())
	require.Equal(t, rules.NewCategoryId("REPETITIONS_STYLE"), r.GetCategory().GetID())
	require.Equal(t, rules.ITSStyle, r.GetLocQualityIssueType())
	require.Equal(t, "GERMAN_WORD_REPEAT_BEGINNING_RULE", r.GetID())
	require.Equal(t, 2, r.MinToCheckParagraph())
	require.NotEmpty(t, r.GetIncorrectExamples())
}
