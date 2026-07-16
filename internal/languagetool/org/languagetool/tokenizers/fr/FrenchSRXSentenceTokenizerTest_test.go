package fr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFrenchSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewFrenchSRXSentenceTokenizer()
	require.NotEmpty(t, tok.Tokenize("A. B."))
}
