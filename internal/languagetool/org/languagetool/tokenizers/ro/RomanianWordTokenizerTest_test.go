package ro

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func tokStr(tokens []string) string {
	return "[" + strings.Join(tokens, ", ") + "]"
}

func TestRomanianWordTokenizer_Tokenize(t *testing.T) {
	w := NewRomanianWordTokenizer()
	assertTok := func(input string, size int, expected string) {
		t.Helper()
		got := w.Tokenize(input)
		require.Equal(t, size, len(got), "input=%q got=%v", input, got)
		require.Equal(t, expected, tokStr(got), "input=%q", input)
	}

	assertTok("Aceaste mese sunt bune", 7, "[Aceaste,  , mese,  , sunt,  , bune]")
	assertTok("Această carte este frumoasă", 7, "[Această,  , carte,  , este,  , frumoasă]")
	assertTok("nu-ți doresc", 5, "[nu, -, ți,  , doresc]")
	assertTok("zicea „merge", 4, "[zicea,  , „, merge]")
	assertTok("zicea „ merge", 5, "[zicea,  , „,  , merge]")
	assertTok("zicea merge”", 4, "[zicea,  , merge, ”]")
	assertTok("zicea „merge bine”", 7, "[zicea,  , „, merge,  , bine, ”]")
	assertTok("ți-am", 3, "[ți, -, am]")
	assertTok("zicea «merge bine»", 7, "[zicea,  , «, merge,  , bine, »]")
	assertTok("zicea <<merge bine>>", 9, "[zicea,  , <, <, merge,  , bine, >, >]")
	assertTok("avea 15% apă", 6, "[avea,  , 15, %,  , apă]")
	assertTok("are 30°C", 5, "[are,  , 30, °, C]")
	assertTok("fructe=mere", 3, "[fructe, =, mere]")
	assertTok("pere|mere", 3, "[pere, |, mere]")
	assertTok("pere\nmere", 3, "[pere, \n, mere]")
	assertTok("pere\rmere", 3, "[pere, \r, mere]")
	assertTok("pere\n\rmere", 4, "[pere, \n, \r, mere]")
	got := w.Tokenize("www.LanguageTool.org")
	require.Equal(t, 1, len(got))
}
