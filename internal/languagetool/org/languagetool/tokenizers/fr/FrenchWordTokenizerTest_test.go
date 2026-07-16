package fr

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func tokStr(tokens []string) string {
	return "[" + strings.Join(tokens, ", ") + "]"
}

func TestFrenchWordTokenizer_Tokenize(t *testing.T) {
	w := NewFrenchWordTokenizer()
	require.Equal(t, 1, len(w.Tokenize("name@example.com")))
	require.Equal(t, 2, len(w.Tokenize("name@example.com.")))
	require.Equal(t, 2, len(w.Tokenize("name@example.com:")))
	require.Equal(t, 7, len(w.Tokenize("L'origen de name@example.com.")))
	require.Equal(t, 4, len(w.Tokenize("jusqu'au bout")))
	require.Equal(t, 2, len(w.Tokenize("d’aujourd’hui")))
	require.Equal(t, 2, len(w.Tokenize("d'aujourd’hui")))
	require.Equal(t, 2, len(w.Tokenize("d'aujourd'hui")))
	require.Equal(t, 1, len(w.Tokenize("entr'ouvrions")))
	require.Equal(t, 1, len(w.Tokenize("entr’ouvrions")))
	require.Equal(t, 2, len(w.Tokenize("Penses-tu")))
	require.Equal(t, 1, len(w.Tokenize("Strauss-Kahn")))
	require.Equal(t, 2, len(w.Tokenize("Semble-t-elle")))
	require.Equal(t, 3, len(w.Tokenize("N’est-il")))
	require.Equal(t, 3, len(w.Tokenize("Faites-le-moi")))
	require.Equal(t, 2, len(w.Tokenize("donne-t-on")))
	require.Equal(t, 3, len(w.Tokenize("qu'est-ce")))
	require.Equal(t, 3, len(w.Tokenize("t'es-tu")))
	require.Equal(t, 1, len(w.Tokenize("rendez-vous")))
	require.Equal(t, 2, len(w.Tokenize("Petit-déjeunes-tu")))
	require.Equal(t, 4, len(w.Tokenize("Y-a-t-il")))
	require.Equal(t, 4, len(w.Tokenize("va-t-en")))
	require.Equal(t, 3, len(w.Tokenize("va-t'en")))
	require.Equal(t, 3, len(w.Tokenize("va-t’en")))
	require.Equal(t, 2, len(w.Tokenize("d'1")))
	require.Equal(t, 1, len(w.Tokenize("Rendez-Vous")))
	require.Equal(t, 1, len(w.Tokenize("sous-trai\u00ADtants")))

	require.Equal(t, "[-, L', homme, .]", tokStr(w.Tokenize("-L'homme.")))
	require.Equal(t, "[-, Oui,  , -, l', homme, .]", tokStr(w.Tokenize("-Oui -l'homme.")))

	require.Equal(t, "[Qu’, est, -ce,  , que,  , ç’, a,  , à,  , voir,  , ?]",
		tokStr(w.Tokenize("Qu’est-ce que ç’a à voir ?")))
	require.Equal(t, "[Qu’, est, -ce,  , que,  , ç', a,  , à,  , voir,  , ?]",
		tokStr(w.Tokenize("Qu’est-ce que ç'a à voir ?")))
	require.Equal(t, "[Ç’, allait,  , être,  , le,  , rêve,  , du,  , XVIIIe,  , siècle, .]",
		tokStr(w.Tokenize("Ç’allait être le rêve du XVIIIe siècle.")))

	require.Equal(t, 1, len(w.Tokenize("10 000")))
	require.Equal(t, 1, len(w.Tokenize("1 000 000")))
	require.Equal(t, 3, len(w.Tokenize("2005 57 114")))
	require.Equal(t, "[2005,  , 57 114]", tokStr(w.Tokenize("2005 57 114")))
	require.Equal(t, 3, len(w.Tokenize("2005 454")))
	require.Equal(t, "[2005,  , 454]", tokStr(w.Tokenize("2005 454")))
	require.Equal(t, 1, len(w.Tokenize("$1")))
	require.Equal(t, "[$1]", tokStr(w.Tokenize("$1")))

	require.Equal(t,
		"[Mais,  , ça,  , ne,  , l', empêche,  , pas,  , de,  , la,  , «,  , décomposer,  , »,  , avec,  , humour, .]",
		tokStr(w.Tokenize("Mais ça ne l'empêche pas de la « décomposer » avec humour.")))
}
