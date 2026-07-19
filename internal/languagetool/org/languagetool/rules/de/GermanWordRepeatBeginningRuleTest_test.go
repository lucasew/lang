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

	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Er ist nett. Er heißt Max."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Außerdem kommt er. Ferner kommt sie. Außerdem kommt es."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("2011: Dieses passiert. 2011: Jenes passiert. 2011: Nicht passiert"))))
	require.Equal(t, 1, len(rule.MatchList(languagetool.SplitAndAnalyze("Er ist nett. Er heißt Max. Er ist 11."))))
	require.Equal(t, 1, len(rule.MatchList(languagetool.SplitAndAnalyze("Außerdem kommt er. Außerdem kommt sie."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Außerdem ist das ein neuer Text."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Außerdem ist das ein neuer Text\n\nAußerdem noch mehr ohne Punkt\n\nAußerdem schon wieder"))))
}

func TestGermanWordRepeatBeginningRule_Category(t *testing.T) {
	r := NewGermanWordRepeatBeginningRule(deWRBMessages())
	require.NotNil(t, r.GetCategory())
	require.Equal(t, rules.NewCategoryId("REPETITIONS_STYLE"), r.GetCategory().GetID())
	require.Equal(t, rules.ITSStyle, r.GetLocQualityIssueType())
}
