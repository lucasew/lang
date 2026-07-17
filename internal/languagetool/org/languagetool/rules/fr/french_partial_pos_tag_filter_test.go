package fr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFrenchPartialPosTagFilter_Accept(t *testing.T) {
	f := NewFrenchPartialPosTagFilter(func(partial string) []string {
		if partial == "chat" {
			return []string{"Ncms"}
		}
		if partial == "manger" {
			return []string{"Vmn"}
		}
		return nil
	})
	ok, err := f.Accept("chatons", "^(chat)ons$", "Nc.*", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = f.Accept("chatons", "^(chat)ons$", "Vm.*", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)

	// negate
	ok, err = f.Accept("chatons", "^(chat)ons$", "Vm.*", true, false, "", "")
	require.NoError(t, err)
	require.True(t, ok) // has tags, none match Vm → negate keeps
}
