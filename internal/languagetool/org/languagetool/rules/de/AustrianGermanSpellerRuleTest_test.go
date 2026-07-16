package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAustrianGermanSpellerRule_Rule(t *testing.T) {
	r := NewAustrianGermanSpellerRule(nil)
	require.Equal(t, "AUSTRIAN_GERMAN_SPELLER_RULE", r.GetID())
	require.False(t, r.IsMisspelled("Haus"))
}
