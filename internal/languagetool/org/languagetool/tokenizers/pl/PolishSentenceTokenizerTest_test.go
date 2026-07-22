package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/tokenizers/pl/PolishSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer = new SRXSentenceTokenizer(new Polish())  // short code "pl"
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitPL mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitPL(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewPolishSRXSentenceTokenizer()
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of PolishSentenceTokenizerTest.testTokenize — all cases, exact equality.
func TestPolishSentenceTokenizer_Tokenize(t *testing.T) {
	testSplitPL(t, "To się wydarzyło 3.10.2000 i mam na to dowody.")

	testSplitPL(t, "To było 13.12 - nikt nie zapomni tego przemówienia.")
	testSplitPL(t, "Heute ist der 13.12.2004.")
	testSplitPL(t, "To jest np. ten debil spod jedynki.")
	testSplitPL(t, "To jest 1. wydanie.")
	testSplitPL(t, "Dziś jest 13. rocznica powstania wąchockiego.")

	testSplitPL(t, "Das in Punkt 3.9.1 genannte Verhalten.")

	testSplitPL(t, "To jest tzw. premier.")
	testSplitPL(t, "Jarek kupił sobie kurteczkę, tj. strój Marka.")

	testSplitPL(t, "„Prezydent jest niemądry”. ", "Tak wyszło.")
	testSplitPL(t, "„Prezydent jest niemądry”, powiedział premier")

	// from user bug reports:
	testSplitPL(t, "Temperatura wody w systemie wynosi 30°C.",
		"W skład obiegu otwartego wchodzi zbiornik i armatura.")
	testSplitPL(t, "Zabudowano kolumny o długości 45 m. ",
		"Woda z ujęcia jest dostarczana do zakładu.")

	// two-letter initials:
	testSplitPL(t, "Najlepszym polskim reżyserem był St. Różewicz. ", "Chodzi o brata wielkiego poety.")

	// From the abbreviation list:
	testSplitPL(t, "Ks. Jankowski jest prof. teologii.")
	testSplitPL(t, "To wydarzyło się w 1939 r.",
		"To był burzliwy rok.")
	testSplitPL(t, "Prezydent jest popierany przez 20 proc. społeczeństwa.")
	testSplitPL(t, "Moje wystąpienie ma na celu zmobilizowanie zarządu partii do działań, które umożliwią uzyskanie 40 proc.",
		"Nie widzę dziś na scenie politycznej formacji, która lepiej by łączyła różne poglądy")
	testSplitPL(t, "To jest zmienna A.", "Zaś to jest zmienna B.")
	// SKROTY_BEZ_KROPKI in ENDABREVLIST
	testSplitPL(t, "Mam już 20 mln.", "To powinno mi wystarczyć")
	testSplitPL(t, "Mam już 20 mln. buraków.")
	// ellipsis
	testSplitPL(t, "Rytmem tej wiecznie przemijającej światowej egzystencji […] rytmem mesjańskiej natury jest szczęście.")
	// sic!
	testSplitPL(t, "W gazecie napisali, że pasy (sic!) pogryzły człowieka.")
	// Numbers with dots.
	testSplitPL(t, "Mam w magazynie dwie skrzynie LMD20. ", "Jestem żołnierzem i wiem, jak można ich użyć")
}
