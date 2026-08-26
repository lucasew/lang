package tools

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.tools.DiffsAsMatchesTest.

// Java-diff-utils SPLIT_BY_WORD_PATTERN uses Pattern \\s without UNICODE_CHARACTER_CLASS.
func TestDiffsAsMatches_NBSPNotWordSplit(t *testing.T) {
	d := NewDiffsAsMatches()
	// NBSP between words is not a delimiter in Java \\s; keep as one token for split.
	// Deleting "a" in "x\u00a0a" still produces a match; positions must stay UTF-16.
	matches := d.GetPseudoMatches("x\u00a0a y", "x\u00a0a z")
	require.NotEmpty(t, matches)
	// "y" → "z" at end; FromPos/ToPos UTF-16: "x\u00a0a " = 4 units, "y" at 4..5
	require.Equal(t, 4, matches[0].GetFromPos())
	require.Equal(t, 5, matches[0].GetToPos())
}

func TestDiffsAsMatches_DiffsAsMatches(t *testing.T) {
	d := NewDiffsAsMatches()

	matches0 := d.GetPseudoMatches(
		`This is a "thing". This is.`,
		`This is a "thing." This is.`,
	)
	require.Len(t, matches0, 1)
	require.Equal(t, `["thing."]`, fmt.Sprint(matches0[0].GetReplacements()))
	require.Equal(t, 10, matches0[0].GetFromPos())
	require.Equal(t, 18, matches0[0].GetToPos())

	matches := d.GetPseudoMatches(
		"This are a sentence with too mistakes.",
		"This is a sentence with two mistakes.",
	)
	require.Len(t, matches, 2)
	require.Equal(t, "[is]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 5, matches[0].GetFromPos())
	require.Equal(t, 8, matches[0].GetToPos())
	require.Equal(t, "[two]", fmt.Sprint(matches[1].GetReplacements()))
	require.Equal(t, 25, matches[1].GetFromPos())
	require.Equal(t, 28, matches[1].GetToPos())

	matches = d.GetPseudoMatches(
		"I am going to er remove one word.",
		"I am going to remove one word.",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 14, matches[0].GetFromPos())
	require.Equal(t, 17, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"And I am going to remove one word.",
		"I am going to remove one word.",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 4, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"I am going to add word.",
		"I am going to add one word.",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[add one]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 14, matches[0].GetFromPos())
	require.Equal(t, 17, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"a word at the start.",
		"Add a word at the start.",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[Add a]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 1, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"Add word at position 1.",
		"Add a word at position 1.",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[Add a]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 3, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"Esta serealiza cada semana.",
		"Esta se realiza cada semana.",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[se realiza]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 5, matches[0].GetFromPos())
	require.Equal(t, 14, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"Una cosa,una altra.",
		"Una cosa, una altra.",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[cosa, ]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 4, matches[0].GetFromPos())
	require.Equal(t, 9, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"Que el año nuevo empezó.",
		"El año nuevo empezó.",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[El]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 6, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"¡Holà ! Estamos aquí.",
		"¡Hola! Estamos aquí.",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[¡Hola!]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 7, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"I truely reserve.",
		"I truly deserve.",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[truly deserve]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 2, matches[0].GetFromPos())
	require.Equal(t, 16, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"}Describes how plants are important",
		"It describes how plants are important",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[It describes]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 10, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"(Calle)Hace falta tener en cuenta una economía estable, un trabajo bueno.",
		"Hace falta tener en cuenta una economía estable, un trabajo bueno.",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 7, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"(Calle)Hace falta tener en cuenta una economía estable, un trabajo bueno.",
		"(Calle) Hace falta tener en cuenta una economía estable, un trabajo bueno.",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[Calle) ]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 1, matches[0].GetFromPos())
	require.Equal(t, 7, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"Joan Caprí fou un actor humorista i monologuista català.",
		"Joan Caprí fou un actor, humorista i monologuista català.",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[actor,]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 18, matches[0].GetFromPos())
	require.Equal(t, 23, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"Ei he vist el teu amic!",
		"Ei, he vist el teu amic!",
	)
	require.Len(t, matches, 1)
	require.Equal(t, "[Ei,]", fmt.Sprint(matches[0].GetReplacements()))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 2, matches[0].GetToPos())

	matches = d.GetPseudoMatches(
		"Va arribar tard a l'examen, perdent així després tota oportunitat d'aprovar l'assignatura.",
		"Va arribar tard a l'examen i, per això, va perdre tota oportunitat d'aprovar l'assignatura.",
	)
	require.Len(t, matches, 1)
}
