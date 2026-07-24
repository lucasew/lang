package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCasesForPreposition(t *testing.T) {
	c := CasesForPreposition("mit")
	require.Contains(t, c, CaseDat)
	c = CasesForPreposition("an")
	require.Contains(t, c, CaseDat)
	require.Contains(t, c, CaseAkk)
	require.Nil(t, CasesForPreposition("xyzzy"))
}
