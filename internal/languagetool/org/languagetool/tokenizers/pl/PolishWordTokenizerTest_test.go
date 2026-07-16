package pl

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

func tokStr(tokens []string) string {
	return "[" + strings.Join(tokens, ", ") + "]"
}

func TestPolishWordTokenizer_Tokenize(t *testing.T) {
	w := NewPolishWordTokenizer()
	tokens := w.Tokenize("To jest\u00A0 test")
	require.Equal(t, 6, len(tokens))
	require.Equal(t, "[To,  , jest, \u00A0,  , test]", tokStr(tokens))

	tokens2 := w.Tokenize("To\rłamie")
	require.Equal(t, 3, len(tokens2))
	require.Equal(t, "[To, \r, łamie]", tokStr(tokens2))

	tokens3 := w.Tokenize("A to jest-naprawdę-test!")
	require.Equal(t, 6, len(tokens3))
	require.Equal(t, "[A,  , to,  , jest-naprawdę-test, !]", tokStr(tokens3))

	tokens4 := w.Tokenize("Niemiecko- i angielsko-polski")
	require.Equal(t, 6, len(tokens4))
	require.Equal(t, "[Niemiecko, -,  , i,  , angielsko-polski]", tokStr(tokens4))

	tokens5 := w.Tokenize("Widzę krowę -i to dobrze!")
	require.Equal(t, 11, len(tokens5))
	require.Equal(t, "[Widzę,  , krowę,  , -, i,  , to,  , dobrze, !]", tokStr(tokens5))

	tokens6 := w.Tokenize("A to jest zdanie—rzeczywiście—z wtrąceniem.")
	require.Equal(t, 14, len(tokens6))
	require.Equal(t, "[A,  , to,  , jest,  , zdanie, —, rzeczywiście, —, z,  , wtrąceniem, .]", tokStr(tokens6))

	// without tagger: compounds stay whole
	compoundSentence := "To jest kobieta-wojownik w polsko-czeskim ubraniu, która wysłała dwa SMS-y."
	compoundTokens := w.Tokenize(compoundSentence)
	require.Equal(t, 21, len(compoundTokens))
	require.Equal(t, "[To,  , jest,  , kobieta-wojownik,  , w,  , polsko-czeskim,  , ubraniu, ,,  , która,  , wysłała,  , dwa,  , SMS-y, .]", tokStr(compoundTokens))
}

func TestPolishWordTokenizer_TokenizeWithTagger(t *testing.T) {
	// Needs Polish BaseTagger / PoliMorfologik for hybrid hyphen compounds.
	tools.Unimplemented("PolishWordTokenizerTest.testTokenize (tagger-dependent compounds)")
}
