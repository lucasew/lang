package da

// Twin of languagetool-language-modules/da/src/test/java/org/languagetool/tokenizers/da/DanishSRXSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer = new SRXSentenceTokenizer(new Danish())  // short code "da"
// No setSingleLineBreaksMarksParagraph — use SRX defaults for Danish.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitDA mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitDA(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewDanishSRXSentenceTokenizer()
	// default paragraph mode — do NOT invent flags unless Java sets them
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of DanishSRXSentenceTokenizerTest.testTokenize — all active cases, exact equality.
func TestDanishSRXSentenceTokenizer_Tokenize(t *testing.T) {
	// NOTE: sentences here need to end with a space character so they
	// have correct whitespace when appended:
	testSplitDA(t, "Dette er en sætning.")
	testSplitDA(t, "Dette er en sætning. ", "Her er den næste.")
	testSplitDA(t, "En sætning! ", "Yderlige en.")
	testSplitDA(t, "En sætning... ", "Yderlige en.")
	testSplitDA(t, "På hjemmesiden http://www.stavekontrolden.dk bygger vi stavekontrollen.")
	testSplitDA(t, "Den 31.12. går ikke!")
	testSplitDA(t, "Den 3.12.2011 går ikke!")
	testSplitDA(t, "I det 18. og tidlige 19. århundrede hentede amerikansk kunst det meste af sin inspiration fra Europa.")

	testSplitDA(t, "Hendes Majestæt Dronning Margrethe II (Margrethe Alexandrine Þórhildur Ingrid, Danmarks dronning) (født 16. april 1940 på Amalienborg Slot) er siden 14. januar 1972 Danmarks regent.")
	testSplitDA(t, "Hun har residensbolig i Christian IX's Palæ på Amalienborg Slot.")
	testSplitDA(t, "Tronfølgeren ledte herefter statsrådsmøderne under Kong Frederik 9.'s fravær.")
	testSplitDA(t, "Marie Hvidt, Frederik IV - En letsindig alvorsmand, Gads Forlag, 2004.")
	testSplitDA(t, "Da vi første gang besøgte Restaurant Chr. IV, var vi de eneste gæster.")

	testSplitDA(t, "I dag er det den 25.12.2010.")
	testSplitDA(t, "I dag er det d. 25.12.2010.")
	testSplitDA(t, "I dag er den 13. december.")
	testSplitDA(t, "Arrangementet starter ca. 17:30 i dag.")
	testSplitDA(t, "Arrangementet starter ca. 17:30.")
	testSplitDA(t, "Det er nævnt i punkt 3.6.4 Rygbelastende helkropsvibrationer.")

	testSplitDA(t, "Rent praktisk er det også lettest lige at mødes, så der kan udveksles nøgler og brugsanvisninger etc.")
	testSplitDA(t, "Andre partier incl. borgerlige partier har deres særlige problemer: nogle samarbejder med apartheidstyret i Sydafrika, med NATO-landet Tyrkiet etc., men det skal så sandelig ikke begrunde en SF-offensiv for et samarbejde med et parti.")

	testSplitDA(t, "Hvad nu,, den bliver også.")
	testSplitDA(t, "Det her er det.. ", "Og her fortsætter det.")

	testSplitDA(t, "Dette er en(!) sætning.")
	testSplitDA(t, "Dette er en(!!) sætning.")
	testSplitDA(t, "Dette er en(?) sætning.")
	testSplitDA(t, "Dette er en(??) sætning.")
	testSplitDA(t, "Dette er en(???) sætning.")
	testSplitDA(t, "Militær værnepligt blev indført (traktaten krævede, at den tyske hær ikke oversteg 100.000 mand).")

	testSplitDA(t, "Siden illustrerede hun \"Historierne om Regnar Lodbrog\" 1979 og \"Bjarkemål\" 1982 samt Poul Ørums \"Komedie i Florens\" 1990.")
}
