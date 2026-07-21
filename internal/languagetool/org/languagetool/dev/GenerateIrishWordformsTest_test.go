package dev

// Twin of GenerateIrishWordformsTest (Java king).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateIrishWordforms_GetEndingsRegex(t *testing.T) {
	// Java test map (subset) — order is longest-first
	m := map[string][]string{
		"óir":   {"m"},
		"eoir":  {"m"},
		"eálaí": {"m"},
	}
	require.Equal(t, "(.+)(eálaí|eoir|óir)$", GetEndingsRegex(m))
}

func TestGenerateIrishWordforms_GuessIrishFSTNounClassSimple(t *testing.T) {
	require.Equal(t, "Nm3-1", GuessIrishFSTNounClassSimple("blagadóir"))
}

func TestGenerateIrishWordforms_ExtractEnWiktionaryNounTemplate(t *testing.T) {
	a := "bádóirí, type: {{ga-decl-m3|b|ádóir|ádóra|ádóirí}}"
	aMap := ExtractEnWiktionaryNounTemplate(a)
	require.Equal(t, "b", aMap["stem"])
	require.Equal(t, "ádóirí", aMap["pl.gen"])
}
