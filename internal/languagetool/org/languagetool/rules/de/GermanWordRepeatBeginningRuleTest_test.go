package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanWordRepeatBeginningRuleTest.java
//
// Java uses JLanguageTool(de-DE) + lt.check with DE_REPEATEDWORDS disabled.
// This twin exercises GermanWordRepeatBeginningRule alone via MatchList +
// SplitAndAnalyze (rule-only outcomes; same counts as Java for this textlevel rule).
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

// Java GermanWordRepeatBeginningRuleTest.testRule — full matrix.
func TestGermanWordRepeatBeginningRule_Rule(t *testing.T) {
	rule := NewGermanWordRepeatBeginningRule(deWRBMessages())

	// correct sentences:
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Er ist nett. Er heißt Max."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Außerdem kommt er. Ferner kommt sie. Außerdem kommt es."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("2011: Dieses passiert. 2011: Jenes passiert. 2011: Nicht passiert"))))
	// errors:
	require.Equal(t, 1, len(rule.MatchList(languagetool.SplitAndAnalyze("Er ist nett. Er heißt Max. Er ist 11."))))
	require.Equal(t, 1, len(rule.MatchList(languagetool.SplitAndAnalyze("Außerdem kommt er. Außerdem kommt sie."))))
	// this used to cause false alarms because reset() was not implemented
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Außerdem ist das ein neuer Text."))))
	// only consider 'real' sentences that end in [.!?]:
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Außerdem ist das ein neuer Text\n\nAußerdem noch mehr ohne Punkt\n\nAußerdem schon wieder"))))
}

// Java ADVERBS set — every surface + case-sensitive negatives.
func TestGermanWordRepeatBeginningRule_AdverbList(t *testing.T) {
	r := NewGermanWordRepeatBeginningRule(deWRBMessages())
	// Exact Java ADVERBS surfaces (incl. Ü in Überdies)
	for _, w := range []string{
		"Auch", "Anschließend", "Außerdem", "Danach", "Ferner",
		"Nebenher", "Nebenbei", "Überdies", "Weiterführend", "Zudem", "Zusätzlich",
	} {
		require.True(t, r.isAdverb(atrWithPOS(w, "ADV", "")), "isAdverb(%q)", w)
	}
	// Not in set / case-sensitive (Java HashSet.contains on getToken surface)
	require.False(t, r.isAdverb(atrWithPOS("Dann", "ADV", "dann")))
	require.False(t, r.isAdverb(atrWithPOS("außerdem", "ADV", "außerdem")))
	require.False(t, r.isAdverb(atrWithPOS("AUSSERDEM", "ADV", "")))
	require.False(t, r.isAdverb(atrWithPOS("uberdies", "ADV", ""))) // missing Ü
	require.False(t, r.isAdverb(nil))
	// set size matches Java (11)
	require.Equal(t, 11, len(germanWordRepeatBeginningAdverbs))
}

// Word vs adverb shortMsg arms from core WordRepeatBeginningRule.match (DE isAdverb).
func TestGermanWordRepeatBeginningRule_WordVsAdverbShortMsg(t *testing.T) {
	rule := NewGermanWordRepeatBeginningRule(deWRBMessages())

	// three successive "Er" → word short message (not adverb)
	ms := rule.MatchList(languagetool.SplitAndAnalyze("Er ist nett. Er heißt Max. Er ist 11."))
	require.Equal(t, 1, len(ms))
	require.Equal(t, deWRBMessages()["desc_repetition_beginning_word"], ms[0].GetShortMessage())
	require.Contains(t, ms[0].GetMessage(), deWRBMessages()["desc_repetition_beginning_thesaurus"])
	// match length UTF-16 of "Er"
	require.Equal(t, 2, ms[0].GetToPos()-ms[0].GetFromPos())

	// two successive "Außerdem" → adverb short message (second sentence)
	ms = rule.MatchList(languagetool.SplitAndAnalyze("Außerdem kommt er. Außerdem kommt sie."))
	require.Equal(t, 1, len(ms))
	require.Equal(t, deWRBMessages()["desc_repetition_beginning_adv"], ms[0].GetShortMessage())
	require.Contains(t, ms[0].GetMessage(), deWRBMessages()["desc_repetition_beginning_thesaurus"])
	// "Außerdem" UTF-16 length (all BMP)
	require.Equal(t, len([]rune("Außerdem")), ms[0].GetToPos()-ms[0].GetFromPos())
}

// Metadata from Java ctor + core WordRepeatBeginningRule inheritance.
func TestGermanWordRepeatBeginningRule_Category(t *testing.T) {
	r := NewGermanWordRepeatBeginningRule(deWRBMessages())
	require.NotNil(t, r.GetCategory())
	require.Equal(t, rules.NewCategoryId("REPETITIONS_STYLE"), r.GetCategory().GetID())
	require.Equal(t, rules.ITSStyle, r.GetLocQualityIssueType())
	require.Equal(t, "GERMAN_WORD_REPEAT_BEGINNING_RULE", r.GetID())
	require.Equal(t, 2, r.MinToCheckParagraph())
	require.Len(t, r.GetIncorrectExamples(), 1)
	require.Equal(t,
		"Dann hatten wir Freizeit. Dann gab es Essen. <marker>Dann</marker> gingen wir schlafen.",
		r.GetIncorrectExamples()[0].GetExample(),
	)
	require.Equal(t, []string{"Schließlich"}, r.GetIncorrectExamples()[0].GetCorrections())
	require.Len(t, r.GetCorrectExamples(), 1)
	require.Equal(t,
		"Dann hatten wir Freizeit. Danach gab es Essen. <marker>Schließlich</marker> gingen wir schlafen.",
		r.GetCorrectExamples()[0].GetExample(),
	)
}
