package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPersonaNumberFromPX(t *testing.T) {
	// PX1MS0S0-style: index 2 = person, 6 = number — use a long enough tag
	p, n := PersonaNumberFromPX("PX1MS0S0")
	require.Equal(t, "1", p)
	require.Equal(t, "S", n)
}

func TestPossessiusRedundantsFilter_PronounFound(t *testing.T) {
	f := NewPossessiusRedundantsFilter()
	got := f.Suggest(PossessiveSuggestionInput{
		PronounFound:     true,
		ApostropheNeeded: true,
		NounToken:        "amic",
	})
	require.Equal(t, "l'amic", got)

	got = f.Suggest(PossessiveSuggestionInput{PronounFound: true, ApostropheNeeded: false})
	require.Equal(t, "", got)
}

func TestPossessiusRedundantsFilter_AddDative(t *testing.T) {
	f := NewPossessiusRedundantsFilter()
	got := f.Suggest(PossessiveSuggestionInput{
		Persona: "3", Number: "S",
		HasSomePronoun:   false,
		VerbToken:        "trenca",
		AroundPossessive: []string{"el", "braç"},
	})
	require.Contains(t, got, "trenca")
	require.Contains(t, got, "li")
}
