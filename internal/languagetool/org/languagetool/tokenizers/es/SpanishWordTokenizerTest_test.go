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

	// Java ORDINAL_POINT leading \b (UNICODE_CHARACTER_CLASS): digit run must
	// not follow a Unicode word char. Non-ASCII letters are word chars in Java
	// but not under Go ASCII \b — must not over-protect/merge.
	tokens = w.Tokenize("ñ1.o")
	require.Equal(t, "[ñ1, ., o]", tokStr(tokens))

	tokens = w.Tokenize("á1.º")
	require.Equal(t, "[á1, ., º]", tokStr(tokens))

	// Java UCC \w includes Mn/Me/Mc — combining mark after suffix is not \b.
	// Must not over-protect (Java: [1, ., ó] with o+U+0301 as one word run).
	tokens = w.Tokenize("1.o\u0301")
	require.Equal(t, "[1, ., o\u0301]", tokStr(tokens))

	// Join_Control (ZWJ/ZWNJ) are UCC \w — trailing \b fails.
	tokens = w.Tokenize("1.o\u200D")
	require.Equal(t, "[1, ., o, \u200D]", tokStr(tokens))

	tokens = w.Tokenize("1.o\u200C")
	require.Equal(t, "[1, ., o, \u200C]", tokStr(tokens))

	// VS-16 (U+FE0F) is Mn → UCC \w; no ORDINAL protect; VS merges onto prior token.
	tokens = w.Tokenize("1.o\uFE0F")
	require.Equal(t, "[1, ., o\uFE0F]", tokStr(tokens))

	// Pc (Connector_Punctuation) beyond ASCII '_' is UCC \w — e.g. undertie U+203F.
	tokens = w.Tokenize("1.o\u203F")
	require.Equal(t, "[1, ., o, \u203F]", tokStr(tokens))

	// ASCII '_' neighbor (Pc) still blocks trailing \b.
	tokens = w.Tokenize("1.o_x")
	require.Equal(t, "[1, ., o_x]", tokStr(tokens))

	// Java UCC \d matches Nd (Arabic-Indic etc.). ORDINAL_POINT protects the
	// point; tokenizer wordCharacters still use ASCII \d so non-ASCII digits
	// stay outside the word run — Java yields [١, .º] / [٢, .a].
	tokens = w.Tokenize("١.º")
	require.Equal(t, "[١, .º]", tokStr(tokens))

	tokens = w.Tokenize("٢.a")
	require.Equal(t, "[٢, .a]", tokStr(tokens))

	tokens = w.Tokenize("al-Ándalus")
	require.Equal(t, 1, len(tokens))
}
