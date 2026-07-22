package sk

// Twin of languagetool-language-modules/sk/src/test/java/org/languagetool/tokenizers/sk/SlovakSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer: setSingleLineBreaksMarksParagraph(true)
// stokenizer2: setSingleLineBreaksMarksParagraph(false)
// private testSplit(...) → TestTools.testSplit(sentences, stokenizer2)
// Skip only Java-commented “known to fail” cases.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitSK mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer2)
// with stokenizer2.setSingleLineBreaksMarksParagraph(false).
func testSplitSK(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewSlovakSRXSentenceTokenizer()
	tok.SetSingleLineBreaksMarksParagraph(false)
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// testSplitSK1 mirrors TestTools.testSplit(sentences, stokenizer)
// with stokenizer.setSingleLineBreaksMarksParagraph(true).
func testSplitSK1(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewSlovakSRXSentenceTokenizer()
	tok.SetSingleLineBreaksMarksParagraph(true)
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of SlovakSentenceTokenizerTest.testTokenize — all active cases, exact equality.
func TestSlovakSentenceTokenizer_Tokenize(t *testing.T) {
	testSplitSK(t, "This is a sentence. ")

	// NOTE: sentences here need to end with a space character so they
	// have correct whitespace when appended:
	testSplitSK(t, "Dies ist ein Satz.")
	testSplitSK(t, "Dies ist ein Satz. ", "Noch einer.")
	testSplitSK(t, "Ein Satz! ", "Noch einer.")
	testSplitSK(t, "Ein Satz... ", "Noch einer.")
	testSplitSK(t, "Unter http://www.test.de gibt es eine Website.")

	testSplitSK(t, "Das ist,, also ob es bla.")
	testSplitSK(t, "Das ist es.. ", "So geht es weiter.")

	testSplitSK(t, "Das hier ist ein(!) Satz.")
	testSplitSK(t, "Das hier ist ein(!!) Satz.")
	testSplitSK(t, "Das hier ist ein(?) Satz.")
	testSplitSK(t, "Das hier ist ein(???) Satz.")
	testSplitSK(t, "Das hier ist ein(???) Satz.")

	// ganzer Satz kommt oder nicht:
	testSplitSK(t, "Das war es: gar nichts.")
	testSplitSK(t, "Das war es: Dies ist ein neuer Satz.")

	// incomplete sentences, need to work for on-thy-fly checking of texts:
	testSplitSK(t, "Here's a")
	testSplitSK(t, "Here's a sentence. ",
		"And here's one that's not comp")

	testSplitSK(t, "„Prezydent jest niemądry”. ", "Tak wyszło.")
	testSplitSK(t, "„Prezydent jest niemądry”, powiedział premier")

	testSplitSK(t, "Das Schreiben ist auf den 3.10. datiert.")
	testSplitSK(t, "Das Schreiben ist auf den 31.1. datiert.")
	testSplitSK(t, "Das Schreiben ist auf den 3.10.2000 datiert.")
	testSplitSK(t, "Toto 2. vydanie bolo rozobrané za 1,5 roka.")
	testSplitSK(t, "Festival Bažant Pohoda slávi svoje 10. výročie.")
	testSplitSK(t, "Dlho odkladané parlamentné voľby v Angole sa uskutočnia 5. septembra.")
	testSplitSK(t, "Das in Punkt 3.9.1 genannte Verhalten.")

	// From the abbreviation list:
	testSplitSK(t, "Aké sú skutočné príčiny tzv. transformačných príznakov?")
	testSplitSK(t, "Aké príplatky zamestnancovi (napr. za nadčas) stanovuje Zákonník práce?")
	testSplitSK(t, "Počas neprítomnosti zastupuje MUDr. Marianna Krupšová.")
	testSplitSK(t, "Staroveký Egypt vznikol okolo r. 3150 p.n.l. (tzn. 3150 pred Kr.). ",
		"A zanikol v r. 31 pr. Kr.")

	// from user bug reports:
	testSplitSK(t, "Temperatura wody w systemie wynosi 30°C.",
		"W skład obiegu otwartego wchodzi zbiornik i armatura.")
	testSplitSK(t, "Zabudowano kolumny o długości 45 m. ",
		"Woda z ujęcia jest dostarczana do zakładu.")

	// two-letter initials:
	testSplitSK(t, "Najlepszym polskim reżyserem był St. Różewicz. ", "Chodzi o brata wielkiego poety.")
	testSplitSK(t, "Nore M. hrozí za podvod 10 až 15 rokov.")
	testSplitSK(t, "To jest zmienna A.", "Zaś to jest zmienna B.")
	// Numbers with dots.
	testSplitSK(t, "Mam w magazynie dwie skrzynie LMD20. ", "Jestem żołnierzem i wiem, jak można ich użyć")
	// ellipsis
	testSplitSK(t, "Rytmem tej wiecznie przemijającej światowej egzystencji […] rytmem mesjańskiej natury jest szczęście.")

	// Tests taken from LanguageTool's SentenceSplitterTest.py:
	testSplitSK(t, "This is a sentence. ")
	testSplitSK(t, "This is a sentence. ", "And this is another one.")
	testSplitSK(t, "This is a sentence.", "Isn't it?", "Yes, it is.")

	testSplitSK(t, "Don't split strings like U. S. A. either.")
	testSplitSK(t, "Don't split strings like U.S.A. either.")
	testSplitSK(t, "Don't split... ", "Well you know. ",
		"Here comes more text.")
	testSplitSK(t, "Don't split... well you know. ",
		"Here comes more text.")
	testSplitSK(t, "The \".\" should not be a delimiter in quotes.")
	testSplitSK(t, "\"Here he comes!\" she said.")
	testSplitSK(t, "\"Here he comes!\", she said.")
	testSplitSK(t, "\"Here he comes.\" ",
		"But this is another sentence.")
	testSplitSK(t, "\"Here he comes!\". ", "That's what he said.")
	testSplitSK(t, "The sentence ends here. ", "(Another sentence.)")
	// known to fail:
	// testSplit(new String[]{"He won't. ", "Really."});
	testSplitSK(t, "He won't go. ", "Really.")
	testSplitSK(t, "He won't say no.", "Not really.")
	testSplitSK(t, "This is it: a test.")
	// one/two returns = paragraph = new sentence:
	// TestTools.testSplit(new String[] { "He won't\n\n", "Really." }, stokenizer2);
	testSplitSK(t, "He won't\n\n", "Really.")
	// TestTools.testSplit(new String[] { "He won't\n", "Really." }, stokenizer);
	testSplitSK1(t, "He won't\n", "Really.")
	// TestTools.testSplit(new String[] { "He won't\n\n", "Really." }, stokenizer2);
	testSplitSK(t, "He won't\n\n", "Really.")
	// TestTools.testSplit(new String[] { "He won't\nReally." }, stokenizer2);
	testSplitSK(t, "He won't\nReally.")
	// Missing space after sentence end:
	testSplitSK(t, "James is from the Ireland!",
		"He lives in Spain now.")
}
