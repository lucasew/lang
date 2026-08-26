package ro

// Twin of languagetool-language-modules/ro/src/test/java/org/languagetool/tokenizers/ro/RomanianSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer: setSingleLineBreaksMarksParagraph(true)
// stokenizer2: setSingleLineBreaksMarksParagraph(false)
// private testSplit(...) → TestTools.testSplit(sentences, stokenizer2)
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitRO mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer2)
// with stokenizer2.setSingleLineBreaksMarksParagraph(false).
func testSplitRO(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewRomanianSRXSentenceTokenizer()
	tok.SetSingleLineBreaksMarksParagraph(false)
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// testSplitRO1 mirrors TestTools.testSplit(sentences, stokenizer)
// with stokenizer.setSingleLineBreaksMarksParagraph(true).
func testSplitRO1(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewRomanianSRXSentenceTokenizer()
	tok.SetSingleLineBreaksMarksParagraph(true)
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of RomanianSentenceTokenizerTest.testTokenize — all active cases, exact equality.
func TestRomanianSentenceTokenizer_Tokenize(t *testing.T) {
	testSplitRO(t, "Aceasta este o propozitie fara diacritice. ")
	testSplitRO(t, "Aceasta este o fraza fara diacritice. ",
		"Propozitia a doua, tot fara diacritice. ")
	testSplitRO(t, "Aceasta este o propoziție cu diacritice. ")
	testSplitRO(t, "Aceasta este o propoziție cu diacritice. ",
		"Propoziția a doua, cu diacritice. ")
	testSplitRO(t, "O propoziție! ",
		"Și încă o propoziție. ")
	testSplitRO(t, "O propoziție... ",
		"Și încă o propoziție. ")
	testSplitRO(t, "La adresa http://www.archeus.ro găsiți resurse lingvistice. ")
	testSplitRO(t, "Data de 10.02.2009 nu trebuie să fie separator de propoziții. ")
	testSplitRO(t, "Astăzi suntem în data de 07.05.2007. ")
	testSplitRO(t, "Astăzi suntem în data de 07/05/2007. ")
	testSplitRO(t, "La anumărul (1) avem puține informații. ")
	testSplitRO(t, "To jest 1. wydanie.")
	testSplitRO(t, "La anumărul 1. avem puține informații. ")
	testSplitRO(t, "La anumărul 13. avem puține informații. ")
	testSplitRO(t, "La anumărul 1.3.3 avem puține informații. ")
	testSplitRO(t, "O singură propoziție... ")
	testSplitRO(t, "Colegii mei s-au dus... ")
	testSplitRO(t, "O singură propoziție!!! ")
	testSplitRO(t, "O singură propoziție??? ")
	testSplitRO(t, "Propoziții: una și alta. ")
	testSplitRO(t, "Domnu' a plecat. ")
	testSplitRO(t, "Profu' de istorie tre' să predea lecția. ")
	testSplitRO(t, "Sal'tare! ")
	testSplitRO(t, "'Neaţa! ")
	testSplitRO(t, "Deodat'apare un urs. ")
	testSplitRO(t, "A făcut două cópii. ")
	testSplitRO(t, "Ionel adúnă acum ceea ce Maria aduná înainte să vin eu. ")
	testSplitRO(t, "Domnu' a plecat")
	testSplitRO(t, "Domnu' a plecat. ",
		"El nu a plecat")
	testSplitRO(t, "Se pot întâlni și abrevieri precum S.U.A. sau B.C.R. într-o singură propoziție.")
	testSplitRO(t, "Se pot întâlni și abrevieri precum S.U.A. sau B.C.R. ",
		"Aici sunt două propoziții.")
	testSplitRO(t, "Același lucru aici... ",
		"Aici sunt două propoziții.")
	testSplitRO(t, "Același lucru aici... dar cu o singură propoziție.")
	testSplitRO(t, "„O propoziție!” ",
		"O alta.")
	testSplitRO(t, "„O propoziție!!!” ",
		"O alta.")
	testSplitRO(t, "„O propoziție?” ",
		"O alta.")
	testSplitRO(t, "„O propoziție?!?” ",
		"O alta.")
	testSplitRO(t, "«O propoziție!» ",
		"O alta.")
	testSplitRO(t, "«O propoziție!!!» ",
		"O alta.")
	testSplitRO(t, "«O propoziție?» ",
		"O alta.")
	testSplitRO(t, "«O propoziție???» ",
		"O alta.")
	testSplitRO(t, "«O propoziție?!?» ",
		"O alta.")
	testSplitRO(t, "O primă propoziție. ",
		"(O alta.)")
	testSplitRO(t, "A venit domnu' Vasile. ")
	testSplitRO(t, "A venit domnu' acela. ")
	testSplitRO(t, "A venit domnul\n\n",
		"Vasile.")
	testSplitRO1(t, "A venit domnul\n",
		"Vasile.")
	testSplitRO(t, "A venit domnu'\n\n",
		"Vasile.")
	testSplitRO1(t, "A venit domnu'\n",
		"Vasile.")
	testSplitRO(t, "El este din România!",
		"Acum e plecat cu afaceri.")
	testSplitRO(t, "Temperatura este de 30°C.",
		"Este destul de cald.")
	testSplitRO(t, "A alergat 50 m. ",
		"Deja a obosit.")
	testSplitRO(t, "Pentru dvs. vom face o excepție.")
	testSplitRO(t, "Pt. dumneavoastră vom face o excepție.")
	testSplitRO(t, "Pt. dvs. vom face o excepție.")
	testSplitRO(t, "A expus problema d.p.d.v. artistic.")
	testSplitRO(t, "A expus problema dpdv. artistic.")
	testSplitRO(t, "Are mere, pere, șamd. dar nu are alune.")
	testSplitRO(t, "Are mere, pere, ș.a.m.d. dar nu are alune.")
	testSplitRO(t, "Are mere, pere, ș.a.m.d. ",
		"În schimb, nu are alune.")
	testSplitRO(t, "Are mere, pere, ş.c.l. dar nu are alune.")
	testSplitRO(t, "Are mere, pere, ş.c.l. ",
		"Nu are alune.")
	testSplitRO(t, "Are mere, pere, etc. dar nu are alune.")
	testSplitRO(t, "Are mere, pere, etc. ",
		"Nu are alune.")
	testSplitRO(t, "Are mere, pere, ș.a. dar nu are alune.")
	testSplitRO(t, "Lecția începe la pag. următoare și are trei pagini.")
	testSplitRO(t, "Lecția începe la pag. 20 și are trei pagini.")
	testSplitRO(t, "A acționat în conformitate cu lg. 144, art. 33.")
	testSplitRO(t, "A acționat în conformitate cu leg. 144, art. 33.")
	testSplitRO(t, "A acționat în conformitate cu legea nr. 11.")
	testSplitRO(t, "Lupta a avut loc în anul 2000 î.H. și a durat trei ani.")
	testSplitRO(t, "Discuția a avut loc pe data de douăzeci aug. și a durat două ore.")
	testSplitRO(t, "Discuția a avut loc pe data de douăzeci ian. și a durat două ore.")
	testSplitRO(t, "Discuția a avut loc pe data de douăzeci feb. și a durat două ore.")
	testSplitRO(t, "Discuția a avut loc pe data de douăzeci ian.",
		"A durat două ore.")
	testSplitRO(t, "A fost și la M.Ap.N. dar nu l-au primit. ")
	testSplitRO(t, "A fost și la M.Ap.N. ",
		"Nu l-au primit. ")
	testSplitRO(t, "Apo' da' tulai (sic!) că mult mai e de mers.")
	testSplitRO(t, "Apo' da' tulai(sic!) că mult mai e de mers.")
	testSplitRO(t, "Aici este o frază […] mult prescurtată.")
	testSplitRO(t, "Aici este o frază [...] mult prescurtată.")
}
