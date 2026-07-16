package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSwissGermanSpellerRule_Rule(t *testing.T) {
	r := NewSwissGermanSpellerRule(nil)
	require.Equal(t, "SWISS_GERMAN_SPELLER_RULE", r.GetID())
}
