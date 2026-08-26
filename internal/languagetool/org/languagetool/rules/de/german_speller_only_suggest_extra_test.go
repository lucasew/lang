package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOnlySuggestions_JavaSwitchFilled(t *testing.T) {
	r := NewGermanSpellerRule(nil)
	// pairs filled from Java getOnlySuggestions switch (were missing before)
	require.Equal(t, []string{"wie viel"}, r.OnlySuggestions("wieviel"))
	require.Equal(t, []string{"an sich"}, r.OnlySuggestions("ansich"))
	require.Equal(t, []string{"Spaß"}, r.OnlySuggestions("Spass"))
	require.Equal(t, []string{"Maßnahme"}, r.OnlySuggestions("Massnahme"))
	require.Equal(t, []string{"Trilogie"}, r.OnlySuggestions("Triologie"))
	require.Equal(t, []string{"immer noch"}, r.OnlySuggestions("immernoch"))
}
