package eo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.tokenizers.eo.EsperantoWordTokenizerTest.

func tokStr(tokens []string) string {
	return "[" + strings.Join(tokens, ", ") + "]"
}

func TestEsperantoWordTokenizer_Tokenize(t *testing.T) {
	w := NewEsperantoWordTokenizer()
	testList := w.Tokenize("Tio estas\u00A0testo")
	require.Equal(t, 5, len(testList))
	require.Equal(t, "[Tio,  , estas, \u00A0, testo]", tokStr(testList))

	testList = w.Tokenize("dank' al 'tio'")
	require.Equal(t, 7, len(testList))
	require.Equal(t, "[dank',  , al,  , ', tio, ']", tokStr(testList))
}
