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
}
