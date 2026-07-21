package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/tokenizers/de/GermanSRXSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func testSplitDE(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewGermanSRXSentenceTokenizer()
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

func TestGermanSRXSentenceTokenizer_Tokenize(t *testing.T) {
	// NOTE (Java): sentences here need to end with a space so they have correct
	// whitespace when appended:
	testSplitDE(t, "Dies ist ein Satz.")
	testSplitDE(t, "Dies ist ein Satz. ", "Noch einer.")
	testSplitDE(t, "Dies ist ein Satz.¹ ", "Noch einer.")
	testSplitDE(t, "Ein Satz! ", "Noch einer.")
	testSplitDE(t, "Ein Satz... ", "Noch einer.")
	testSplitDE(t, "Unter http://www.test.de gibt es eine Website.")
	testSplitDE(t, "Das Schreiben ist auf den 3.10. datiert.")
	testSplitDE(t, "Das Schreiben ist auf den 31.1. datiert.")
	testSplitDE(t, "Das Schreiben ist auf den 3.10.2000 datiert.")
	testSplitDE(t, "Natürliche Vererbungsprozesse prägten sich erst im 18. und frühen 19. Jahrhundert aus.")
	testSplitDE(t, "Das ist ja 1a. ", "Und das auch.")
	testSplitDE(t, "Hallo, ich bin’s. ", "Könntest du kommen?")
	testSplitDE(t, "In der 1. Bundesliga kam es zum Eklat.")
	testSplitDE(t, "Dies ist, z. B., ein Satz.")
	testSplitDE(t, "Da hatte es über 30 °C. ", "Hier kommt der nächste Satz.")

	testSplitDE(t, "Das 1. Internationale Filmfestival findet nächste Woche statt.")
	testSplitDE(t, "Friedrich I., auch bekannt als Friedrich der Große.")
	testSplitDE(t, "Friedrich II., auch bekannt als Friedrich der Große.")
	testSplitDE(t, "Friedrich IIXC., auch bekannt als Friedrich der Große.")
	testSplitDE(t, "Friedrich II. öfter auch bekannt als Friedrich der Große.")
	testSplitDE(t, "Friedrich VII. öfter auch bekannt als Friedrich der Große.")
	testSplitDE(t, "Friedrich X. öfter auch bekannt als Friedrich der Zehnte.")

	// non-breaking space (HTML editors)
	tok := NewGermanSRXSentenceTokenizer()
	require.Equal(t, 2, len(tok.Tokenize("Dies ist ein Satz. \u00A0Noch einer.")))
	require.Equal(t, 2, len(tok.Tokenize("Dies ist ein Satz.   \u00A0Noch einer.")))
	require.Equal(t, 2, len(tok.Tokenize("Dies ist ein Satz.\u00A0 Noch einer.")))
	require.Equal(t, 2, len(tok.Tokenize("Dies ist ein Satz.\u00A0\u00A0\u00A0 Noch einer.")))
	testSplitDE(t, "Ein Satz!\u00A0", "Noch einer.")
	testSplitDE(t, "Dies ist, z.\u00A0B., ein Satz.")
	testSplitDE(t, "Hier steht was mit mehr Wörtern, weil wir mal sehen wollen, wie denn so die Erkennung der Satzlänge geht, wenn die Sätze doch deutlich länger werden, also wirklich deutlich länger als das normal.\u00A0", "Hier steht etwas anderes.")

	testSplitDE(t, "Heute ist der 13.12.2004.")
	testSplitDE(t, "Heute ist der 13. Dezember.")
	testSplitDE(t, "Heute ist der 1. Januar.")
	testSplitDE(t, "Es geht am 24.09. los.")
	testSplitDE(t, "Es geht um ca. 17:00 los.")
	testSplitDE(t, "Das in Punkt 3.9.1 genannte Verhalten.")

	testSplitDE(t, "Diese Periode begann im 13. Jahrhundert und damit bla.")
	testSplitDE(t, "Diese Periode begann im 13. oder 14. Jahrhundert und damit bla.")
	testSplitDE(t, "Diese Periode datiert auf das 13. bis zum 14. Jahrhundert und damit bla.")

	testSplitDE(t, "Das gilt lt. aktuellem Plan.")
	testSplitDE(t, "Orangen, Äpfel etc. werden gekauft.")

	testSplitDE(t, "Das ist,, also ob es bla.")
	testSplitDE(t, "Das ist es.. ", "So geht es weiter.")

	testSplitDE(t, "Das hier ist ein(!) Satz.")
	testSplitDE(t, "Das hier ist ein(!!) Satz.")
	testSplitDE(t, "Das hier ist ein(?) Satz.")
	testSplitDE(t, "Das hier ist ein(???) Satz.")

	testSplitDE(t, "»Der Papagei ist grün.« ", "Das kam so.")
	testSplitDE(t, "»Der Papagei ist grün«, sagte er")

	// colon: Java currently does not distinguish sentence vs not
	testSplitDE(t, "Das war es: gar nichts.")
	testSplitDE(t, "Das war es: Dies ist ein neuer Satz.")

	// Crime and Punishment regression cases
	testSplitDE(t, "schlug er die Richtung nach der K … brücke ein. ")
	testSplitDE(t, "sobald ich es von einem Freunde zurückbekomme …« Er wurde verlegen und schwieg.")
	testSplitDE(t, "Er kannte eine Unmenge Quellen, aus denen er schöpfen konnte, d. h. natürlich, wo er durch Arbeit sich etwas verdienen konnte.")
	testSplitDE(t, "Stimme am lautesten heraustönte …. ", "Sobald er auf der Straße war")
	testSplitDE(t, "»Welche Wohnung?\" ", "»Die, wo wir arbeiten.")
	testSplitDE(t, "»Nun also, wie ist's?« fragte Lushin und blickte sie fest an.")
	testSplitDE(t, "»Nun also, wie ist es?« fragte Lushin und blickte sie fest an.")
	testSplitDE(t, "„Nun also, wie ist es?“ fragte Lushin und blickte sie fest an.")
	testSplitDE(t, "»Nun also, wie ist es?« ", "Dann ging er.")

	testSplitDE(t, "Dies ist ein Satz mit einer EMail.Addresse@example.com!")
	testSplitDE(t, "Sonderbarerweise sind auch Beispiel!Eins@example.com und Foo?Bar@example.com valide.")

	testSplitDE(t, "Er kauft Obst, Gemüse, Brot, Milch, usw. ", "Danach geht er nach Hause.")
	testSplitDE(t, "Sie kaufte Bücher, Hefte, Stifte usw., weil sie sich auf das neue Schuljahr vorbereiten wollte.")
	testSplitDE(t, "Im Supermarkt gibt es viele Sorten von Käse, Wurst, Brot usw. und alle sind sehr lecker.")

	testSplitDE(t, "Im Büro gibt es Computer, Drucker, Telefone etc. ", "Die Mitarbeiter arbeiten den ganzen Tag an verschiedenen Projekten.")
	testSplitDE(t, "Die Ergebnisse sind in Tabellen, Grafiken, Diagrammen etc. dargestellt.")

	testSplitDE(t, "Master of Law (LL.B.)")
	testSplitDE(t, "Master of Law (LL. B.)")
}

// Abbreviations that block UPPERCASE_SENTENCE_START false positives (Java JLT SRX).
func TestGermanSRXSentenceTokenizer_AbbrevNoSplit(t *testing.T) {
	tok := NewGermanSRXSentenceTokenizer()
	for _, s := range []string{
		"Dieser Satz ist bspw. okay so.",
		"Dieser Satz ist z.B. okay so.",
		"Dies ist, z. B., ein Satz.",
	} {
		require.Equal(t, 1, len(tok.Tokenize(s)), s)
	}
}
