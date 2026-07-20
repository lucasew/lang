package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/tokenizers/pt/PortugueseWordTokenizerTest.java
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func testTokenise(t *testing.T, sentence string, tokens ...string) {
	t.Helper()
	// Java keeps dictionary-tagged hyphen compounds via PortugueseTagger.
	// Inject IsTaggedPT for those surfaces — no soft invent doNotSplit lexicon.
	prev := IsTaggedPT
	IsTaggedPT = func(s string) bool {
		switch strings.ToLower(s) {
		// Java PortugueseWordTokenizerTest hyphen compounds kept whole by tagger
		case "sex-appeal", "aix-en-provence", "montemor-o-novo", "andorra-a-velha", "tsé-tung",
			"jiu-jitsu", "franco-prussiano",
			"diz-se", "amamo-lo", "fi-lo", "pusé-lo", "canta-lo", "dar-no-lo",
			"fá-lo-á", "dir-lhe-ia", "banhar-nos-emos",
			"soto-pôr", "soto-trepar":
			return true
		default:
			return false
		}
	}
	t.Cleanup(func() { IsTaggedPT = prev })

	w := NewPortugueseWordTokenizer()
	got := w.Tokenize(sentence)
	require.Equal(t, tokens, got, "tokenize(%q)", sentence)
}

func TestPortugueseWordTokenizer_TokeniseBreakChars(t *testing.T) {
	testTokenise(t, "Isto é\u00A0um teste", "Isto", " ", "é", "\u00A0", "um", " ", "teste")
	testTokenise(t, "Isto\rquebra", "Isto", "\r", "quebra")
}

func TestPortugueseWordTokenizer_TokeniseHyphenNoWhiteSpace(t *testing.T) {
	testTokenise(t, "Agora isto sim é-mesmo!-um teste.",
		"Agora", " ", "isto", " ", "sim", " ", "é", "-", "mesmo", "!", "-", "um", " ", "teste", ".")
}

func TestPortugueseWordTokenizer_TokeniseWordFinalHyphen(t *testing.T) {
	testTokenise(t, "Agora isto é- realmente!- um teste.",
		"Agora", " ", "isto", " ", "é", "-", " ", "realmente", "!", "-", " ", "um", " ", "teste", ".")
}

func TestPortugueseWordTokenizer_TokeniseMDash(t *testing.T) {
	testTokenise(t, "Agora isto é—realmente!—um teste.",
		"Agora", " ", "isto", " ", "é", "—", "realmente", "!", "—", "um", " ", "teste", ".")
}

func TestPortugueseWordTokenizer_TokeniseHyphenatedSingleToken(t *testing.T) {
	testTokenise(t, "sex-appeal", "sex-appeal")
	testTokenise(t, "Aix-en-Provence", "Aix-en-Provence")
	testTokenise(t, "Montemor-o-Novo", "Montemor-o-Novo")
	testTokenise(t, "Andorra-a-Velha", "Andorra-a-Velha")
	testTokenise(t, "Tsé-Tung", "Tsé-Tung")
}

func TestPortugueseWordTokenizer_TokeniseHyphenatedSplitRegardlessOfLetterCase(t *testing.T) {
	testTokenise(t, "jiu-jitsu", "jiu-jitsu")
	testTokenise(t, "Jiu-jitsu", "Jiu-jitsu")
	testTokenise(t, "JIU-JITSU", "JIU-JITSU")
	testTokenise(t, "Jiu-Jitsu", "Jiu-Jitsu")
	testTokenise(t, "franco-prussiano", "franco-prussiano")
	testTokenise(t, "Franco-prussiano", "Franco-prussiano")
	testTokenise(t, "Franco-Prussiano", "Franco-Prussiano")
}

func TestPortugueseWordTokenizer_TokeniseHyphenatedSplit(t *testing.T) {
	testTokenise(t, "Paris-São Paulo", "Paris", "-", "São", " ", "Paulo")
	testTokenise(t, "Sem-Peixe", "Sem", "-", "Peixe")
	testTokenise(t, "húngaro-americano", "húngaro", "-", "americano")
}

func TestPortugueseWordTokenizer_TokeniseHyphenatedClitics(t *testing.T) {
	testTokenise(t, "diz-se", "diz-se")
	testTokenise(t, "amamo-lo", "amamo-lo")
	testTokenise(t, "fi-lo", "fi-lo")
	testTokenise(t, "pusé-lo", "pusé-lo")
	testTokenise(t, "canta-lo", "canta-lo")
	testTokenise(t, "dar-no-lo", "dar-no-lo")
	testTokenise(t, "dê-mo", "dê", "-", "mo")
}

func TestPortugueseWordTokenizer_TokeniseMesoclisis(t *testing.T) {
	testTokenise(t, "fá-lo-á", "fá-lo-á")
	testTokenise(t, "dir-lhe-ia", "dir-lhe-ia")
	testTokenise(t, "banhar-nos-emos", "banhar-nos-emos")
}

func TestPortugueseWordTokenizer_TokeniseProductivePrefixes(t *testing.T) {
	testTokenise(t, "soto-pôr", "soto-pôr")
	testTokenise(t, "soto-trepar", "soto-trepar")
}

func TestPortugueseWordTokenizer_TokeniseApostrophe(t *testing.T) {
	testTokenise(t, "d'água", "d", "'", "água")
	testTokenise(t, "d’água", "d", "’", "água")
}

func TestPortugueseWordTokenizer_TokeniseHashtags(t *testing.T) {
	testTokenise(t, "#CantadasDoBem", "#", "CantadasDoBem")
}

func TestPortugueseWordTokenizer_DoNotTokeniseUserMentions(t *testing.T) {
	testTokenise(t, "@user", "@user")
}

func TestPortugueseWordTokenizer_TokeniseCurrency(t *testing.T) {
	testTokenise(t, "R$45,00", "R$", "45,00")
	testTokenise(t, "5£", "5", "£")
	testTokenise(t, "US$249,99", "US$", "249,99")
	testTokenise(t, "€2.000,00", "€", "2.000,00")
}

func TestPortugueseWordTokenizer_TokeniseSplitsPercent(t *testing.T) {
	testTokenise(t, "50%", "50%")
	testTokenise(t, "50%%", "50%", "%")
	testTokenise(t, "50%OFF", "50%", "OFF")
	testTokenise(t, "%50", "%", "50")
	testTokenise(t, "%", "%")
}

func TestPortugueseWordTokenizer_TokeniseNumberAbbreviation(t *testing.T) {
	testTokenise(t, "Nº666", "Nº666")
	testTokenise(t, "N°666", "N°666")
	testTokenise(t, "Nº 420", "Nº", " ", "420")
	testTokenise(t, "N.º69", "N", ".", "º69")
	testTokenise(t, "N.º 80085", "N", ".", "º", " ", "80085")
}

func TestPortugueseWordTokenizer_DoNotTokeniseOrdinalSuperscript(t *testing.T) {
	testTokenise(t, "1º", "1º")
	testTokenise(t, "2.º", "2.º")
	testTokenise(t, "3ºˢ", "3ºˢ")
	testTokenise(t, "4.ºˢ", "4.ºˢ")
	testTokenise(t, "5ª", "5ª")
	testTokenise(t, "6ª", "6ª")
	testTokenise(t, "7ªˢ", "7ªˢ")
	testTokenise(t, "8ªˢ", "8ªˢ")
	testTokenise(t, "9ᵒ", "9ᵒ")
	testTokenise(t, "10.ᵒ", "10.ᵒ")
	testTokenise(t, "11ᵒˢ", "11ᵒˢ")
	testTokenise(t, "12.ᵒˢ", "12.ᵒˢ")
	testTokenise(t, "13ᵃ", "13ᵃ")
	testTokenise(t, "14.ᵃ", "14.ᵃ")
	testTokenise(t, "15ᵃˢ", "15ᵃˢ")
	testTokenise(t, "16.ᵃˢ", "16.ᵃˢ")
	testTokenise(t, "17o", "17o")
	testTokenise(t, "18.o", "18.o")
	testTokenise(t, "19os", "19os")
	testTokenise(t, "20.os", "20.os")
	testTokenise(t, "21a", "21a")
	testTokenise(t, "22.a", "22.a")
	testTokenise(t, "23as", "23as")
	testTokenise(t, "24.as", "24.as")
}

func TestPortugueseWordTokenizer_DoNotTokeniseDegreeExpressions(t *testing.T) {
	testTokenise(t, "25°", "25°")
	testTokenise(t, "26,0°", "26,0°")
	testTokenise(t, "27.0°", "27.0°")
	testTokenise(t, "28,0°C", "28,0°C")
	testTokenise(t, "29.0°C", "29.0°C")
	testTokenise(t, "30,0°c", "30,0°c")
	testTokenise(t, "31.0°c", "31.0°c")
	testTokenise(t, "32°Ra", "32°Ra")
	testTokenise(t, "33,1°Rø", "33,1°Rø")
	testTokenise(t, "34°N", "34°N")
}

func TestPortugueseWordTokenizer_DoNotTokeniseSpaceSeparatedThousands(t *testing.T) {
	testTokenise(t, "35 000", "35 000")
	testTokenise(t, "36 000 000", "36 000 000")
	testTokenise(t, "37 000,00", "37 000,00")
	testTokenise(t, "38 000 000,00", "38 000 000,00")
	testTokenise(t, "39 000°", "39 000°")
	testTokenise(t, "40 000%", "40 000%")
	testTokenise(t, "41 000º", "41 000º")
	testTokenise(t, "42 000o", "42 000o")
	testTokenise(t, "43 00", "43", " ", "00")
}

func TestPortugueseWordTokenizer_TokeniseExponent(t *testing.T) {
	testTokenise(t, "km²", "km", "²")
}

func TestPortugueseWordTokenizer_TokeniseCopyrightAndSimilarSymbols(t *testing.T) {
	testTokenise(t, "Copyright©", "Copyright", "©")
	testTokenise(t, "Bacana®", "Bacana", "®")
	testTokenise(t, "Legal™", "Legal", "™")
}

func TestPortugueseWordTokenizer_TokeniseEmoji(t *testing.T) {
	testTokenise(t, "☺☺☺Só", "☺", "☺", "☺", "Só")
}

func TestPortugueseWordTokenizer_DoNotTokeniseModifierDiacritics(t *testing.T) {
	testTokenise(t, "Não", "Não")
}

func TestPortugueseWordTokenizer_TokeniseExtraWordEdgeChars(t *testing.T) {
	testTokenise(t, "@50", "@50")
	testTokenise(t, "@@50", "@", "@50")
	testTokenise(t, "50@", "50", "@")
	testTokenise(t, "666@50", "666", "@50")
	testTokenise(t, "50‰", "50‰")
	testTokenise(t, "50‰‰", "50‰", "‰")
	testTokenise(t, "‰50", "‰", "50")
	testTokenise(t, "50‰666", "50‰", "666")
}

func TestPortugueseWordTokenizer_TokeniseRarePunctuation(t *testing.T) {
	testTokenise(t, "⌈Herói⌋", "⌈", "Herói", "⌋")
	testTokenise(t, "″Santo Antônio do Manga″", "″", "Santo", " ", "Antônio", " ", "do", " ", "Manga", "″")
}

func TestPortugueseWordTokenizer_TokeniseParagraphSymbol(t *testing.T) {
	testTokenise(t, "§1º", "§", "1º")
}

func TestPortugueseWordTokenizer_TokeniseComplexEmoji(t *testing.T) {
	testTokenise(t, "🧝🏽‍♀️", "🧝", "🏽", "‍", "♀️")
}

func TestPortugueseWordTokenizer_TokeniseUnitsOfMeasure(t *testing.T) {
	testTokenise(t, "100mm", "100mm")
	testTokenise(t, "10x10mm", "10x10mm")
	testTokenise(t, "10×10mm", "10", "×", "10mm")
}
