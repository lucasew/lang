package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/tokenizers/ca/CatalanWordTokenizerTest.java
import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func tokStr(tokens []string) string {
	return "[" + strings.Join(tokens, ", ") + "]"
}

// pronomFeble / proclitic forms CatalanTagger keeps whole after pattern split.
var (
	testPronomFeble = regexp.MustCompile(`(?i)^(['’]en|['’]hi|['’]ho|['’]l|['’]ls|['’]m|['’]n|['’]ns|['’]s|['’]t|-el|-els|-em|-en|-ens|-hi|-ho|-l|-la|-les|-li|-lo|-los|-m|-me|-n|-ne|-nos|-s|-se|-t|-te|-us|-vos)$`)
	testProclitic   = regexp.MustCompile(`(?i)^[lnmtsd]['’]$`)
)

func TestCatalanWordTokenizer_Tokenize(t *testing.T) {
	// Java keeps dictionary-tagged hyphen compounds via CatalanTagger.
	// Inject IsTaggedCA for those surfaces — no soft invent doNotSplit lexicon.
	prev := IsTaggedCA
	IsTaggedCA = func(s string) bool {
		if testPronomFeble.MatchString(s) || testProclitic.MatchString(s) {
			return true
		}
		switch strings.ToLower(s) {
		case "vint-i-quatre", "mont-ras", "emília-romanya", "tel-aviv",
			"abans-d'ahir", "sud-est", "nord-est", "sud-oest", "nord-oest",
			"qui-sap-lo", "qui-sap-la", "qui-sap-los", "qui-sap-les":
			// Sud-Est must split (Title-Title not in dict the same way as Sud-est).
			// Only exact case-insensitive match for lower-style compounds; reject
			// when both sides are Title-case for compass forms.
			if strings.Contains(s, "-") {
				parts := strings.Split(s, "-")
				if len(parts) == 2 {
					low := strings.ToLower(s)
					switch low {
					case "sud-est", "nord-est", "sud-oest", "nord-oest":
						// Keep Sud-est (second lower); split Sud-Est.
						if hasTitleStartCA(parts[0]) && hasTitleStartCA(parts[1]) {
							return false
						}
					}
				}
			}
			return true
		default:
			return false
		}
	}
	t.Cleanup(func() { IsTaggedCA = prev })

	w := INSTANCE

	tokens := w.Tokenize("-contar-se'n-")
	require.Equal(t, "[-, contar, -se, 'n, -]", tokStr(tokens))
	tokens = w.Tokenize("-M'agradaria.")
	require.Equal(t, "[-, M', agradaria, .]", tokStr(tokens))

	tokens = w.Tokenize("Visiteu 'http://www.softcatala.org'")
	require.Equal(t, "[Visiteu,  , ', http://www.softcatala.org, ']", tokStr(tokens))
	tokens = w.Tokenize("Visiteu \"http://www.softcatala.org\"")
	require.Equal(t, "[Visiteu,  , \", http://www.softcatala.org, \"]", tokStr(tokens))
	tokens = w.Tokenize("name@example.com")
	require.Equal(t, 1, len(tokens))
	tokens = w.Tokenize("name@example.com.")
	require.Equal(t, 2, len(tokens))
	tokens = w.Tokenize("name@example.com:")
	require.Equal(t, 2, len(tokens))
	tokens = w.Tokenize("L'origen de name@example.com.")
	require.Equal(t, 7, len(tokens))
	require.Equal(t, "[L', origen,  , de,  , name@example.com, .]", tokStr(tokens))
	tokens = w.Tokenize("L'origen de name@example.com i de name2@example.com.")
	require.Equal(t, 13, len(tokens))
	require.Equal(t, "[L', origen,  , de,  , name@example.com,  , i,  , de,  , name2@example.com, .]", tokStr(tokens))
	tokens = w.Tokenize("L'\"ala bastarda\".")
	require.Equal(t, 7, len(tokens))
	require.Equal(t, "[L', \", ala,  , bastarda, \", .]", tokStr(tokens))
	tokens = w.Tokenize("d'\"ala bastarda\".")
	require.Equal(t, 7, len(tokens))
	require.Equal(t, "[d', \", ala,  , bastarda, \", .]", tokStr(tokens))
	tokens = w.Tokenize("Emporta-te'ls a l'observatori dels mars")
	require.Equal(t, 13, len(tokens))
	require.Equal(t, "[Emporta, -te, 'ls,  , a,  , l', observatori,  , de, ls,  , mars]", tokStr(tokens))
	tokens = w.Tokenize("Emporta-te’ls a l’observatori dels mars")
	require.Equal(t, 13, len(tokens))
	require.Equal(t, "[Emporta, -te, ’ls,  , a,  , l’, observatori,  , de, ls,  , mars]", tokStr(tokens))
	tokens = w.Tokenize("‘El tren Barcelona-València’")
	require.Equal(t, 9, len(tokens))
	require.Equal(t, "[‘, El,  , tren,  , Barcelona, -, València, ’]", tokStr(tokens))
	tokens = w.Tokenize("El tren Barcelona-València")
	require.Equal(t, 7, len(tokens))
	require.Equal(t, "[El,  , tren,  , Barcelona, -, València]", tokStr(tokens))
	tokens = w.Tokenize("No acabava d’entendre’l bé")
	require.Equal(t, 9, len(tokens))
	require.Equal(t, "[No,  , acabava,  , d’, entendre, ’l,  , bé]", tokStr(tokens))
	tokens = w.Tokenize("N'hi ha vint-i-quatre")
	require.Equal(t, 6, len(tokens))
	require.Equal(t, "[N', hi,  , ha,  , vint-i-quatre]", tokStr(tokens))
	tokens = w.Tokenize("Mont-ras")
	require.Equal(t, 1, len(tokens))
	require.Equal(t, "[Mont-ras]", tokStr(tokens))
	tokens = w.Tokenize("És d'1 km.")
	require.Equal(t, 7, len(tokens))
	require.Equal(t, "[És,  , d', 1,  , km, .]", tokStr(tokens))
	tokens = w.Tokenize("És d'1,5 km.")
	require.Equal(t, 7, len(tokens))
	require.Equal(t, "[És,  , d', 1,5,  , km, .]", tokStr(tokens))
	tokens = w.Tokenize("És d'5 km.")
	require.Equal(t, 7, len(tokens))
	require.Equal(t, "[És,  , d', 5,  , km, .]", tokStr(tokens))
	tokens = w.Tokenize("la direcció E-SE")
	require.Equal(t, 7, len(tokens))
	require.Equal(t, "[la,  , direcció,  , E, -, SE]", tokStr(tokens))
	tokens = w.Tokenize("la direcció NW-SE")
	require.Equal(t, 7, len(tokens))
	require.Equal(t, "[la,  , direcció,  , NW, -, SE]", tokStr(tokens))
	tokens = w.Tokenize("Se'n dóna vergonya")
	require.Equal(t, 6, len(tokens))
	require.Equal(t, "[Se, 'n,  , dóna,  , vergonya]", tokStr(tokens))
	tokens = w.Tokenize("Emília-Romanya")
	require.Equal(t, 1, len(tokens))
	require.Equal(t, "[Emília-Romanya]", tokStr(tokens))
	tokens = w.Tokenize("L'Emília-Romanya")
	require.Equal(t, 2, len(tokens))
	require.Equal(t, "[L', Emília-Romanya]", tokStr(tokens))
	tokens = w.Tokenize("col·laboració")
	require.Equal(t, 1, len(tokens))
	tokens = w.Tokenize("col.laboració")
	require.Equal(t, 1, len(tokens))
	tokens = w.Tokenize("col•laboració")
	require.Equal(t, 1, len(tokens))
	tokens = w.Tokenize("col·Laboració")
	require.Equal(t, 1, len(tokens))
	tokens = w.Tokenize("COL∙LABORADORES")
	require.Equal(t, 1, len(tokens))

	tokens = w.Tokenize("abans-d’ahir")
	require.Equal(t, 1, len(tokens))
	tokens = w.Tokenize("abans-d'ahir")
	require.Equal(t, 1, len(tokens))
	tokens = w.Tokenize("Sud-Est")
	require.Equal(t, 3, len(tokens))
	require.Equal(t, "[Sud, -, Est]", tokStr(tokens))
	tokens = w.Tokenize("Sud-est")
	require.Equal(t, 1, len(tokens))
	tokens = w.Tokenize("10 000")
	require.Equal(t, 1, len(tokens))
	tokens = w.Tokenize("1 000 000")
	require.Equal(t, 1, len(tokens))
	tokens = w.Tokenize("2005 57 114")
	require.Equal(t, 3, len(tokens))
	require.Equal(t, "[2005,  , 57 114]", tokStr(tokens))
	tokens = w.Tokenize("2005 454")
	require.Equal(t, 3, len(tokens))
	require.Equal(t, "[2005,  , 454]", tokStr(tokens))
	tokens = w.Tokenize("$1")
	require.Equal(t, 1, len(tokens))
	require.Equal(t, "[$1]", tokStr(tokens))

	tokens = w.Tokenize("AVALUA'T")
	require.Equal(t, 2, len(tokens))
	require.Equal(t, "[AVALUA, 'T]", tokStr(tokens))
	tokens = w.Tokenize("Tel-Aviv")
	require.Equal(t, 1, len(tokens))
	require.Equal(t, "[Tel-Aviv]", tokStr(tokens))
	tokens = w.Tokenize("\"El cas 'Barcelona'\"")
	require.Equal(t, 9, len(tokens))
	require.Equal(t, "[\", El,  , cas,  , ', Barcelona, ', \"]", tokStr(tokens))
	tokens = w.Tokenize("\"El cas 'd'aquell'\"")
	require.Equal(t, 10, len(tokens))
	require.Equal(t, "[\", El,  , cas,  , ', d', aquell, ', \"]", tokStr(tokens))
	tokens = w.Tokenize("\"El cas ‘d’aquell’\"")
	require.Equal(t, 10, len(tokens))
	require.Equal(t, "[\", El,  , cas,  , ‘, d’, aquell, ’, \"]", tokStr(tokens))
	tokens = w.Tokenize("Sàsser-l'Alguer")
	require.Equal(t, 4, len(tokens))
	require.Equal(t, "[Sàsser, -, l', Alguer]", tokStr(tokens))
	tokens = w.Tokenize("Castella-la Manxa")
	require.Equal(t, 5, len(tokens))
	require.Equal(t, "[Castella, -, la,  , Manxa]", tokStr(tokens))
	tokens = w.Tokenize("Qui-sap-lo temps")
	require.Equal(t, 3, len(tokens))
	require.Equal(t, "[Qui-sap-lo,  , temps]", tokStr(tokens))

	tokens = w.Tokenize("Sol Picó (\U0001F40C+\U0001F41A)")
	require.Equal(t, "[Sol,  , Picó,  , (, \U0001F40C, +, \U0001F41A, )]", tokStr(tokens))

	tokens = w.Tokenize("\U0001F9E1prova.")
	require.Equal(t, "[\U0001F9E1, prova, .]", tokStr(tokens))

	tokens = w.Tokenize("\U0001F9E1\U0001F9E1prova\U0001F9E1")
	require.Equal(t, "[\U0001F9E1, \U0001F9E1, prova, \U0001F9E1]", tokStr(tokens))

	tokens = w.Tokenize("❤\uFE0F")
	require.Equal(t, "[❤\uFE0F]", tokStr(tokens))

	tokens = w.Tokenize("❤\uFE0Fprova")
	require.Equal(t, "[❤\uFE0F, prova]", tokStr(tokens))

	tokens = w.Tokenize("H₂O")
	require.Equal(t, "[H₂O]", tokStr(tokens))

	tokens = w.Tokenize("\U0001F9E1")
	require.Equal(t, "[\U0001F9E1]", tokStr(tokens))

	tokens = w.Tokenize("sol∙licitud")
	require.Equal(t, "[sol.licitud]", tokStr(tokens))

	tokens = w.Tokenize("És ell ㅡva dir.")
	require.Equal(t, "[És,  , ell,  , ㅡva,  , dir, .]", tokStr(tokens))
}

func hasTitleStartCA(s string) bool {
	if s == "" {
		return false
	}
	r := []rune(s)[0]
	return (r >= 'A' && r <= 'Z') || (r > 127 && strings.ToUpper(string(r)) == string(r) && strings.ToLower(string(r)) != string(r))
}
