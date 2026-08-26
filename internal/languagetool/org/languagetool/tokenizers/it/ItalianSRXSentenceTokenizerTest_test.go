package it

// Twin of languagetool-language-modules/it/src/test/java/org/languagetool/tokenizers/it/ItalianSRXSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer = new SRXSentenceTokenizer(new Italian())  // short code "it"
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitIT mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitIT(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewItalianSRXSentenceTokenizer()
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of ItalianSRXSentenceTokenizerTest.testTokenize — all non-@Ignore cases, exact equality.
func TestItalianSRXSentenceTokenizer_Tokenize(t *testing.T) {
	testSplitIT(t,
		"Il Castello Reale di Racconigi è situato a Racconigi, in provincia di Cuneo ma poco distante da Torino. ",
		"Nel corso della sua quasi millenaria storia ha visto numerosi rimaneggiamenti e divenne di proprietà dei Savoia a partire dalla seconda metà del XIV secolo.",
	)
	testSplitIT(t, "Dott. Bunsen Honeydew") // abbreviation
	testSplitIT(t,
		"Abbiamo isolato N. meningitidis da un campione di sangue. ",
		"La Prov. di Bolzano ha competenze autonome. ",
		"La Reg. d’Abruzzo confina con il Lazio. ",
		"Il cd. regolamento è stato approvato ieri. ",
		"Alcuni frutti, es. mele e pere, sono disponibili. ",
		"Nel XIX sec. si verificarono grandi cambiamenti. ",
		"Lavora nel sett. energetico da anni. ",
		"La diagnosi è compatibile con sdr. metabolica. ",
		"“Basta!” disse Maria. ",
		"Prima del ricovero, tutti gli accertamenti necessari (esami ematici, ECG, TAC, ecc.) sono stati completati secondo protocollo.",
	)
}
