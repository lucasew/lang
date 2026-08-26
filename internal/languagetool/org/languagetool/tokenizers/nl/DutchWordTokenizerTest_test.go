package nl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.tokenizers.nl.DutchWordTokenizerTest.

func assertTokenize(t *testing.T, input, expected string) {
	t.Helper()
	result := NewDutchWordTokenizer().Tokenize(input)
	require.Equal(t, expected, "["+strings.Join(result, ", ")+"]", "input=%q", input)
}

func TestDutchWordTokenizer_Tokenize(t *testing.T) {
	assertTokenize(t, "This is\u00A0a test", "[This,  , is, \u00A0, a,  , test]")
	assertTokenize(t, "Bla bla oma's bla bla 'test", "[Bla,  , bla,  , oma's,  , bla,  , bla,  , ', test]")
	assertTokenize(t, "Bla bla oma`s bla bla 'test", "[Bla,  , bla,  , oma`s,  , bla,  , bla,  , ', test]")
	assertTokenize(t, "Ik zie het''", "[Ik,  , zie,  , het, ', ']")
	assertTokenize(t, "Ik zie het``", "[Ik,  , zie,  , het, `, `]")
	assertTokenize(t, "''Ik zie het", "[', ', Ik,  , zie,  , het]")

	assertTokenize(t, "Ik 'zie' het", "[Ik,  , ', zie, ',  , het]")
	assertTokenize(t, "Ik ‘zie’ het", "[Ik,  , ‘, zie, ’,  , het]")
	assertTokenize(t, "Ik \"zie\" het", "[Ik,  , \", zie, \",  , het]")
	assertTokenize(t, "Ik “zie” het", "[Ik,  , “, zie, ”,  , het]")
	assertTokenize(t, "'zie'", "[', zie, ']")
	assertTokenize(t, "‘zie’", "[‘, zie, ’]")
	assertTokenize(t, "\"zie\"", "[\", zie, \"]")
	assertTokenize(t, "“zie”", "[“, zie, ”]")

	assertTokenize(t, "Ik `zie het", "[Ik,  , `, zie,  , het]")
	assertTokenize(t, "Ik ``zie het", "[Ik,  , `, `, zie,  , het]")
	assertTokenize(t, "'", "[']")
	assertTokenize(t, "''", "[, ', ']")
	assertTokenize(t, "'x'", "[', x, ']")
	assertTokenize(t, "`x`", "[`, x, `]")
}
