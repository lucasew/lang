package nl

// Twin of languagetool-language-modules/nl/src/test/java/org/languagetool/tokenizers/nl/DutchSRXSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer = new SRXSentenceTokenizer(new Dutch())  // short code "nl"
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitNL mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitNL(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewDutchSRXSentenceTokenizer()
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of DutchSRXSentenceTokenizerTest.testTokenize — all non-@Ignore cases, exact equality.
func TestDutchSRXSentenceTokenizer_Tokenize(t *testing.T) {
	// NOTE: sentences here need to end with a space character so they
	// have correct whitespace when appended:
	testSplitNL(t, "Dit is een zin.")
	testSplitNL(t, "Dit is een zin. ", "Nog een.")
	testSplitNL(t, "Een zin! ", "Nog een.")
	testSplitNL(t, "‘Dat meen je niet!’ kirde Mandy.")
	testSplitNL(t, "Een zin... ", "Nog een.")
	testSplitNL(t, "'En nu.. daden!' aan premier Mark Rutte.")
	testSplitNL(t, "Op http://www.test.de vind je een website.")
	testSplitNL(t, "De brief is op 3-10 gedateerd.")
	testSplitNL(t, "De brief is op 31-1 gedateerd.")
	testSplitNL(t, "De brief is op 3-10-2000 gedateerd.")

	testSplitNL(t, "Vandaag is het 13-12-2004.")
	testSplitNL(t, "Op 24.09 begint het.")
	testSplitNL(t, "Om 17:00 begint het.")
	testSplitNL(t, "In paragraaf 3.9.1 is dat beschreven.")

	testSplitNL(t, "Januari jl. is dat vastgelegd.")
	testSplitNL(t, "Appel en pruimen enz. werden gekocht.")
	testSplitNL(t, "De afkorting n.v.t. betekent niet van toepassing.")

	testSplitNL(t, "Bla et al. blah blah.")

	testSplitNL(t, "Dat is,, of het is bla.")
	testSplitNL(t, "Dat is het.. ", "Zo gaat het verder.")

	testSplitNL(t, "Dit hier is een(!) zin.")
	testSplitNL(t, "Dit hier is een(!!) zin.")
	testSplitNL(t, "Dit hier is een(?) zin.")
	testSplitNL(t, "Dit hier is een(???) zin.")
	testSplitNL(t, "Dit hier is een(???) zin.")

	testSplitNL(t, "Als voetballer wordt hij nooit een prof. ", "Maar prof. N.A.W. Th.Ch. Janssen wordt dat wel.")

	// TODO, zin na dubbele punt
	testSplitNL(t, "Dat was het: helemaal niets.")
	testSplitNL(t, "Dat was het: het is een nieuwe zin.")

	// https://nl.wikipedia.org/wiki/Aanhalingsteken
	testSplitNL(t, "Jan zei: \"Hallo.\"")
	testSplitNL(t, "Jan zei: “Hallo.”")
	testSplitNL(t, "„Hallo,” zei Jan, „waar ga je naartoe?”")
	testSplitNL(t, "„Gisteren”, zei Jan, „was het veel beter weer.”")
	testSplitNL(t, "Jan zei: „Annette zei ‚Hallo’.”")
	testSplitNL(t, "Jan zei: “Annette zei ‘Hallo’.”")
	testSplitNL(t, "Jan zei: «Annette zei ‹Hallo›.»")
	testSplitNL(t, "Wegens zijn „ziekte” hoefde hij niet te werken.")
	testSplitNL(t, "de letter „a”")
	testSplitNL(t, "het woord „beta” is afkomstig van ...")

	// http://taaladvies.net/taal/advies/vraag/11
	testSplitNL(t, "De voorzitter zei: 'Ik zie hier geen been in.'")
	testSplitNL(t, "De voorzitter zei: \"Ik zie hier geen been in.\"")
	testSplitNL(t, "De koning zei: \"Ik herinner me nog dat iemand 'Vive la république' riep tijdens mijn eedaflegging.\"")
	testSplitNL(t, "De koning zei: 'Ik herinner me nog dat iemand \"Vive la république\" riep tijdens mijn eedaflegging.'")
	testSplitNL(t, "De koning zei: 'Ik herinner me nog dat iemand 'Vive la république' riep tijdens mijn eedaflegging.'")
	testSplitNL(t, "Otto dacht: wat een nare verhalen hoor je toch tegenwoordig.")

	// http://taaladvies.net/taal/advies/vraag/871
	testSplitNL(t, "'Ik vrees', zei Rob, 'dat de brug zal instorten.'")
	testSplitNL(t, "'Ieder land', aldus minister Powell, 'moet rekenschap afleggen over de wijze waarop het zijn burgers behandelt.'")
	testSplitNL(t, "'Zeg Rob,' vroeg Jolanda, 'denk jij dat de brug zal instorten?'")
	testSplitNL(t, "'Deze man heeft er niets mee te maken,' aldus korpschef Jan de Vries, 'maar hij heeft momenteel geen leven.'")
	testSplitNL(t, "'Ik vrees,' zei Rob, 'dat de brug zal instorten.'")
	testSplitNL(t, "'Ieder land,' aldus minister Powell, 'moet rekenschap afleggen over de wijze waarop het zijn burgers behandelt.'")
	testSplitNL(t, "'Zeg Rob,' vroeg Jolanda, 'denk jij dat de brug zal instorten?'")
	testSplitNL(t, "'Deze man heeft er niets mee te maken,' aldus korpschef Jan de Vries, 'maar hij heeft momenteel geen leven.'")

	// http://taaladvies.net/taal/advies/vraag/872
	testSplitNL(t, "Zij antwoordde: 'Ik denk niet dat ik nog langer met je om wil gaan.'")
	testSplitNL(t, "Zij fluisterde iets van 'eeuwig trouw' en 'altijd blijven houden van'.")
	testSplitNL(t, "'Heb je dat boek al gelezen?', vroeg hij.")
	testSplitNL(t, "De auteur vermeldt: 'Deze opvatting van het vorstendom heeft lang doorgewerkt.'")

	// http://taaladvies.net/taal/advies/vraag/1557
	testSplitNL(t, "'Gaat u zitten', zei zij. ", "'De dokter komt zo.'")
	testSplitNL(t, "'Mijn broer woont ook in Den Haag', vertelde ze. ", "'Hij woont al een paar jaar samen.'")
	testSplitNL(t, "'Je bent grappig', zei ze. ", "'Echt, ik vind je grappig.'")
	testSplitNL(t, "'Is Jan thuis?', vroeg Piet. ", "'Ik wil hem wat vragen.'")
	testSplitNL(t, "'Ik denk er niet over!', riep ze. ", "'Dat gaat echt te ver, hoor!'")
	testSplitNL(t, "'Ik vermoed', zei Piet, 'dat Jan al wel thuis is.'")

	testSplitNL(t, "Het is een .Net programma. ", "Of een .NEt programma.")
	testSplitNL(t, "Het is een .Net-programma. ", "Of een .NEt-programma.")

	testSplitNL(t, "SP werd in 2001 de sp.a (Socialistische Partij Anders) en heet sinds 2021 Vooruit.")
	testSplitNL(t, "SP.A grijpt terug naar naam met geschiedenis: VOORUIT.")
}
