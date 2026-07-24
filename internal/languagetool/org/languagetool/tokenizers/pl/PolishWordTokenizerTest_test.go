package pl_test

import (
	"strings"
	"testing"

	pltag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pl"
	pl "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/pl"
	"github.com/stretchr/testify/require"
)

func tokStr(tokens []string) string {
	return "[" + strings.Join(tokens, ", ") + "]"
}

// polishTaggerAsHyphen adapts *pltag.PolishTagger to PolishHyphenTagger
// (Java: wordTokenizer.setTagger(pl.getTagger())).
func polishTaggerAsHyphen(t *pltag.PolishTagger) pl.PolishHyphenTagger {
	if t == nil {
		return nil
	}
	return pl.ATRTagFunc(func(tokens []string) []pl.PolishHyphenReadings {
		atrs := t.Tag(tokens)
		out := make([]pl.PolishHyphenReadings, len(atrs))
		for i, a := range atrs {
			out[i] = a
		}
		return out
	})
}

// Twin of org.languagetool.tokenizers.pl.PolishWordTokenizerTest.testTokenize.
func TestPolishWordTokenizer_Tokenize(t *testing.T) {
	wordTokenizer := pl.NewPolishWordTokenizer()
	tokens := wordTokenizer.Tokenize("To jest\u00A0 test")
	require.Equal(t, 6, len(tokens))
	require.Equal(t, "[To,  , jest, \u00A0,  , test]", tokStr(tokens))

	tokens2 := wordTokenizer.Tokenize("To\rłamie")
	require.Equal(t, 3, len(tokens2))
	require.Equal(t, "[To, \r, łamie]", tokStr(tokens2))

	// hyphen with no whitespace
	tokens3 := wordTokenizer.Tokenize("A to jest-naprawdę-test!")
	require.Equal(t, 6, len(tokens3))
	require.Equal(t, "[A,  , to,  , jest-naprawdę-test, !]", tokStr(tokens3))

	// hyphen at the end of the word
	tokens4 := wordTokenizer.Tokenize("Niemiecko- i angielsko-polski")
	require.Equal(t, 6, len(tokens4))
	require.Equal(t, "[Niemiecko, -,  , i,  , angielsko-polski]", tokStr(tokens4))

	// hyphen probably instead of mdash
	tokens5 := wordTokenizer.Tokenize("Widzę krowę -i to dobrze!")
	require.Equal(t, 11, len(tokens5))
	require.Equal(t, "[Widzę,  , krowę,  , -, i,  , to,  , dobrze, !]", tokStr(tokens5))

	// mdash
	tokens6 := wordTokenizer.Tokenize("A to jest zdanie—rzeczywiście—z wtrąceniem.")
	require.Equal(t, 14, len(tokens6))
	require.Equal(t, "[A,  , to,  , jest,  , zdanie, —, rzeczywiście, —, z,  , wtrąceniem, .]", tokStr(tokens6))

	// compound words with hyphens (tagger null — stay whole)
	compoundSentence := "To jest kobieta-wojownik w polsko-czeskim ubraniu, która wysłała dwa SMS-y."
	compoundTokens := wordTokenizer.Tokenize(compoundSentence)
	require.Equal(t, 21, len(compoundTokens))
	require.Equal(t, "[To,  , jest,  , kobieta-wojownik,  , w,  , polsko-czeskim,  , ubraniu, ,,  , która,  , wysłała,  , dwa,  , SMS-y, .]", tokStr(compoundTokens))

	// now setup the tagger... Java: Language pl = new Polish(); wordTokenizer.setTagger(pl.getTagger())
	pltag.EnsureDefaultPolishTagger()
	if pltag.DefaultPolishTagger == nil {
		t.Skip("polish.dict not available; cannot twin setTagger section")
	}
	wordTokenizer.SetTagger(polishTaggerAsHyphen(pltag.DefaultPolishTagger))
	compoundTokens = wordTokenizer.Tokenize(compoundSentence)
	// we should get 4 more tokens: two hyphen tokens and two for the split words
	require.Equal(t, 25, len(compoundTokens))
	require.Equal(t, "[To,  , jest,  , kobieta, -, wojownik,  , w,  , polsko, -, czeskim,  , ubraniu, ,,  , która,  , wysłała,  , dwa,  , SMS-y, .]", tokStr(compoundTokens))

	compoundTokens = wordTokenizer.Tokenize("Miała osiemnaście-dwadzieścia lat.")
	require.Equal(t, 8, len(compoundTokens))
	require.Equal(t, "[Miała,  , osiemnaście, -, dwadzieścia,  , lat, .]", tokStr(compoundTokens))

	// now three-part adja-adja-adj...:
	compoundTokens = wordTokenizer.Tokenize("Słownik polsko-niemiecko-indonezyjski")
	require.Equal(t, 7, len(compoundTokens))
	require.Equal(t, "[Słownik,  , polsko, -, niemiecko, -, indonezyjski]", tokStr(compoundTokens))

	// number ranges:
	compoundTokens = wordTokenizer.Tokenize("Impreza odbędzie się w dniach 1-23 maja.")
	require.Equal(t, 16, len(compoundTokens))
	require.Equal(t, "[Impreza,  , odbędzie,  , się,  , w,  , dniach,  , 1, -, 23,  , maja, .]", tokStr(compoundTokens))

	// number ranges:
	compoundTokens = wordTokenizer.Tokenize("Impreza odbędzie się w dniach 1--23 maja.")
	require.Equal(t, 18, len(compoundTokens))
	require.Equal(t, "[Impreza,  , odbędzie,  , się,  , w,  , dniach,  , 1, -, , -, 23,  , maja, .]", tokStr(compoundTokens))
}
