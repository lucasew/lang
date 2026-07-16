package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikGermanyGermanSpellerRule(t *testing.T) {
	r := NewMorfologikGermanyGermanSpellerRule(nil)
	require.NotNil(t, r)
	require.Equal(t, "GERMAN_SPELLER_RULE", r.GetID())
}
