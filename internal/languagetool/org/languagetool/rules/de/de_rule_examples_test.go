package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Java DashRule / GermanCompoundRule / CompoundInfinitivRule example pairs.
func TestDERule_ExamplePairs(t *testing.T) {
	dash := NewDashRule(nil)
	require.Equal(t, "DE_DASH", dash.GetID())
	require.Contains(t, dash.GetURL(), "grammatik-leerzeichen")
	require.Equal(t, []string{"Diäten-Erhöhung"}, dash.GetIncorrectExamples()[0].GetCorrections())

	comp := NewGermanCompoundRule(nil)
	require.Equal(t, "DE_COMPOUNDS", comp.GetID())
	require.Equal(t, []string{"HNO-Arzt"}, comp.GetIncorrectExamples()[0].GetCorrections())

	inf := NewCompoundInfinitivRule(nil)
	require.Equal(t, "COMPOUND_INFINITIV_RULE", inf.GetID())
	require.Contains(t, inf.GetURL(), "zu-zusammen-oder-getrennt")
	require.Equal(t, []string{"sicherzugehen"}, inf.GetIncorrectExamples()[0].GetCorrections())

	// CaseRule / DuUpperLowerCase / WiederVsWider / word-repeat
	require.Equal(t, []string{"Das Laufen"}, NewCaseRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Du"}, NewDuUpperLowerCaseRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"wider"}, NewWiederVsWiderRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"ist"}, NewGermanWordRepeatRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Schließlich"}, NewGermanWordRepeatBeginningRule(nil).GetIncorrectExamples()[0].GetCorrections())

	// Whitespace / WWIC / punctuation / brackets / grammar
	require.Equal(t, []string{" Das"}, NewSentenceWhitespaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Mine"}, NewGermanWrongWordInContextRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"a. D."}, NewGermanDoublePunctuationRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"("}, NewGermanUnpairedBracketsRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"fehlt"}, NewMissingVerbRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Das Haus"}, NewAgreementRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"bin"}, NewVerbAgreementRule(nil).GetIncorrectExamples()[0].GetCorrections())

	// Quotes / coherency / compounds / names / phrases / style / agreement2 / SVA
	require.Equal(t, []string{"›"}, NewGermanUnpairedQuotesRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Delfine"}, NewWordCoherencyRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Leerzeile"}, NewProhibitedCompoundRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Müller"}, NewSimilarNameRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, "Das ist <marker>allem Anschein nach</marker> eine Phrase.", NewUnnecessaryPhraseRule(nil).GetIncorrectExamples()[0].GetExample())
	require.Contains(t, NewGermanStyleRepeatedWordRule(nil).GetIncorrectExamples()[0].GetExample(), "<marker>gehe</marker>")
	require.Equal(t, []string{"Kleines Haus"}, NewAgreementRule2(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"sind"}, NewSubjectVerbAgreementRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Contains(t, NewStyleRepeatedVeryShortSentences(nil).GetIncorrectExamples()[0].GetExample(), "näher.")
}
