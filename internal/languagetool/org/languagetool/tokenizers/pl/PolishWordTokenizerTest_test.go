package pl

import (
	"strings"
	"testing"

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

	// number ranges split without tagger
	n := w.Tokenize("Impreza odbędzie się w dniach 1-23 maja.")
	require.Equal(t, "[Impreza,  , odbędzie,  , się,  , w,  , dniach,  , 1, -, 23,  , maja, .]", tokStr(n))
}

// mockHyphenTagger marks parts as adjective compounds (adja + adj:).
type mockHyphenTagger struct{}

func (mockHyphenTagger) Tag(tokens []string) []PolishTokenReadings {
	out := make([]PolishTokenReadings, len(tokens))
	// last is full compound — untagged
	for i := 0; i < len(tokens)-1; i++ {
		if i == 0 {
			out[i] = PolishTokenReadings{IsTagged: true, HasAdja: true}
		} else {
			out[i] = PolishTokenReadings{IsTagged: true, HasAdjPartial: true}
		}
	}
	out[len(tokens)-1] = PolishTokenReadings{IsTagged: false}
	// for 3-part: all first as adja except last part adj
	if len(tokens) == 4 { // 3 parts + full
		out[0] = PolishTokenReadings{IsTagged: true, HasAdja: true}
		out[1] = PolishTokenReadings{IsTagged: true, HasAdja: true}
		out[2] = PolishTokenReadings{IsTagged: true, HasAdjPartial: true}
		out[3] = PolishTokenReadings{IsTagged: false}
	}
	// for 2-part subst compounds like kobieta-wojownik — use subst partial
	if len(tokens) == 3 {
		// prefer adj pattern for polsko-czeskim; subst for kobieta-wojownik
		// caller sets tokens; we use subst for any 2-part when first isn't known adja prefix
		out[0] = PolishTokenReadings{IsTagged: true, HasSubstPartial: true}
		out[1] = PolishTokenReadings{IsTagged: true, HasSubstPartial: true}
		out[2] = PolishTokenReadings{IsTagged: false}
	}
	return out
}

// adjHyphenTagger always uses adja/adj for 2-part compounds.
type adjHyphenTagger struct{}

func (adjHyphenTagger) Tag(tokens []string) []PolishTokenReadings {
	out := make([]PolishTokenReadings, len(tokens))
	if len(tokens) == 4 {
		out[0] = PolishTokenReadings{IsTagged: true, HasAdja: true}
		out[1] = PolishTokenReadings{IsTagged: true, HasAdja: true}
		out[2] = PolishTokenReadings{IsTagged: true, HasAdjPartial: true}
		out[3] = PolishTokenReadings{IsTagged: false}
		return out
	}
	if len(tokens) >= 3 {
		out[0] = PolishTokenReadings{IsTagged: true, HasAdja: true}
		out[1] = PolishTokenReadings{IsTagged: true, HasAdjPartial: true}
		out[len(tokens)-1] = PolishTokenReadings{IsTagged: false}
	}
	return out
}

func TestPolishWordTokenizer_TokenizeWithTagger(t *testing.T) {
	w := NewPolishWordTokenizer()
	w.SetTagger(mockHyphenTagger{})
	// 2-part subst compound splits
	got := w.Tokenize("kobieta-wojownik")
	require.Equal(t, "[kobieta, -, wojownik]", tokStr(got))

	w.SetTagger(adjHyphenTagger{})
	got = w.Tokenize("polsko-czeskim")
	require.Equal(t, "[polsko, -, czeskim]", tokStr(got))

	got = w.Tokenize("polsko-niemiecko-indonezyjski")
	require.Equal(t, "[polsko, -, niemiecko, -, indonezyjski]", tokStr(got))
}
