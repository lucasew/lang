package es

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func tokStr(tokens []string) string {
	return "[" + strings.Join(tokens, ", ") + "]"
}

func TestSpanishWordTokenizer_Tokenize(t *testing.T) {
	// Java keeps dictionary-tagged hyphen compounds via SpanishTagger.
	// Inject IsTaggedES for those surfaces — no soft invent exception list.
	prev := IsTaggedES
	IsTaggedES = func(s string) bool {
		switch strings.ToLower(s) {
		case "best-seller", "covid-19", "e-mails", "e-mail", "al-ándalus", "al-andalus":
			return true
		default:
			return false
		}
	}
	t.Cleanup(func() { IsTaggedES = prev })

	w := NewSpanishWordTokenizer()
	tokens := w.Tokenize("*test+")
	require.Equal(t, 3, len(tokens))
	require.Equal(t, "[*, test, +]", tokStr(tokens))

	tokens = w.Tokenize("best-seller Covid-19;sars-cov-2")
	require.Equal(t, 5, len(tokens))
	require.Equal(t, "[best-seller,  , Covid-19, ;, sars-cov-2]", tokStr(tokens))

	tokens = w.Tokenize("e-mails")
	require.Equal(t, 1, len(tokens))
	require.Equal(t, "[e-mails]", tokStr(tokens))

	tokens = w.Tokenize("$100")
	require.Equal(t, 1, len(tokens))
	require.Equal(t, "[$100]", tokStr(tokens))

	tokens = w.Tokenize("$1.000")
	require.Equal(t, 1, len(tokens))
	require.Equal(t, "[$1.000]", tokStr(tokens))

	tokens = w.Tokenize("$1,400.50")
	require.Equal(t, 1, len(tokens))
	require.Equal(t, "[$1,400.50]", tokStr(tokens))

	tokens = w.Tokenize("1,400.50$")
	require.Equal(t, 1, len(tokens))
	require.Equal(t, "[1,400.50$]", tokStr(tokens))

	tokens = w.Tokenize("Ven ‒dijo.") // \u2012
	require.Equal(t, 5, len(tokens))
	require.Equal(t, "[Ven,  , ‒, dijo, .]", tokStr(tokens))

	tokens = w.Tokenize("1.º")
	require.Equal(t, 1, len(tokens))

	tokens = w.Tokenize("Es la 21.ª y el 45.º")
	require.Equal(t, 11, len(tokens))

	tokens = w.Tokenize("Es la 21.a y el 45.o")
	require.Equal(t, 11, len(tokens))

	tokens = w.Tokenize("11.as Jornadas de Estudio")
	require.Equal(t, 7, len(tokens))
	require.Equal(t, "[11.as,  , Jornadas,  , de,  , Estudio]", tokStr(tokens))

	// Java ORDINAL_POINT trailing \b (UNICODE_CHARACTER_CLASS): suffix must be
	// followed by a word boundary. "1.apple" / "1.asiento" must not protect/merge.
	tokens = w.Tokenize("1.apple")
	require.Equal(t, "[1, ., apple]", tokStr(tokens))

	tokens = w.Tokenize("1.asiento")
	require.Equal(t, "[1, ., asiento]", tokStr(tokens))

	tokens = w.Tokenize("21.asiento")
	require.Equal(t, "[21, ., asiento]", tokStr(tokens))

	// Still protect real ordinals when boundary follows (end or non-word).
	tokens = w.Tokenize("1.er.")
	require.Equal(t, "[1.er, .]", tokStr(tokens))

	tokens = w.Tokenize("al-Ándalus")
	require.Equal(t, 1, len(tokens))
}
