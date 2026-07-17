package dev

// Twin of GenerateIrishWordformsTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateIrishWordforms_NoTests(t *testing.T) {
	lines := ExpandIrishNounFromGuess("scríbhneoir")
	require.NotEmpty(t, lines)
}
