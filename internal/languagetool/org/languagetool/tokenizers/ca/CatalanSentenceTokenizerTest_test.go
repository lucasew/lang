package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/tokenizers/ca/CatalanSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer = new SRXSentenceTokenizer(Catalan.getInstance()) — default paragraph mode
// (no setSingleLineBreaksMarksParagraph in this test).
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitCA mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitCA(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewCatalanSRXSentenceTokenizer()
	// match Java Catalan SRX default paragraph mode — do NOT invent flags unless Java sets them
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of CatalanSentenceTokenizerTest.testTokenize — all active cases, exact equality.
func TestCatalanSentenceTokenizer_Tokenize(t *testing.T) {
	// Simple sentences
	testSplitCA(t, "Això és una frase. ", "Això és una altra frase.")
	testSplitCA(t, "Aquesta és l'egua. ", "Aquell és el cavall.")
	testSplitCA(t, "Aquesta és l'egua? ", "Aquell és el cavall.")
	testSplitCA(t, "Vols col·laborar? ", "Sí, i tant.")
	testSplitCA(t, "Com vas d'il·lusió? ", "Bé, bé.")
	testSplitCA(t, "Com vas d’il·lusió? ", "Bé, bé.")
	testSplitCA(t, "És d’abans-d’ahir? ", "Bé, bé.")
	testSplitCA(t, "És d’abans-d’ahir! ", "Bé, bé.")
	testSplitCA(t, "Què vols dir? ", "Ja ho tinc!")
	testSplitCA(t, "Què? ", "Ja ho tinc!")
	testSplitCA(t, "Ah! ", "Ja ho tinc!")
	testSplitCA(t, "Ja ho tinc! ", "Què vols dir?")
	testSplitCA(t, "Us explicaré com va anar: ",
		"»La Maria va engegar el cotxe")
	testSplitCA(t, "diu que va dir. ", "A mi em feia estrany.")
	testSplitCA(t, "Són del s. III dC. ", "Són importants les pintures.")
	testSplitCA(t, "Primera frase.[4] ", "Segona frase")
	testSplitCA(t, "23. Article vint-i-tres")

	// N., t.
	testSplitCA(t, "Vés-te’n. ", "A mi em feia estrany.")
	testSplitCA(t, "Vés-te'n. ", "A mi em feia estrany.")
	testSplitCA(t, "VÉS-TE'N. ", "A mi em feia estrany.")
	testSplitCA(t, "Canten. ", "A mi em feia estrany.")
	testSplitCA(t, "Desprèn. ", "A mi em feia estrany.")
	testSplitCA(t, "(n. 3).")
	testSplitCA(t, " n. 3")
	testSplitCA(t, "n. 3")
	testSplitCA(t, "(\"n. 3\".")
	testSplitCA(t, "En el t. 2 de la col·lecció")
	testSplitCA(t, "Llança't. ", "Fes-ho.")
	testSplitCA(t, "És professor a l'Inst. Joan Vives.")

	// Initials
	testSplitCA(t, "A l'atenció d'A. Comes.")
	testSplitCA(t, "A l'atenció d'À. Comes.")
	testSplitCA(t, "Són els alumnes de Física I. ", "Ara no venen ben preparats.")
	testSplitCA(t, "Va ser obra de Felip V. ", "Ara ho sabem.")
	testSplitCA(t, "Va ser obra d'Alfons X. ", "Ara ho sabem.")
	testSplitCA(t, "Núm. operació 220130000138.")

	// Ellipsis
	testSplitCA(t, "el vi no és gens propi de monjos, amb tot...\" vetllant, això sí")
	testSplitCA(t, "Desenganyeu-vos… ",
		"L’únic problema seriós de l'home en aquest món és el de subsistir.")
	testSplitCA(t, "és clar… traduir és una feina endimoniada")
	testSplitCA(t, "«El cordó del frare…» surt d'una manera desguitarrada")
	testSplitCA(t, "convidar el seu heroi –del ram que sigui–… a prendre cafè.")

	// Abbreviations
	testSplitCA(t, "No Mr. Spock sinó un altre.")
	testSplitCA(t, "Vegeu el cap. 24 del llibre.")
	testSplitCA(t, "Vegeu el cap. IX del llibre.")
	testSplitCA(t, "Viu al núm. 24 del carrer de l'Hort.")
	testSplitCA(t, "Viu al núm. vint-i-quatre del carrer de l'Hort.")
	testSplitCA(t, "El Dr. Joan no vindrà.")
	testSplitCA(t, "Distingit Sr. Joan,")
	testSplitCA(t, "Molt Hble. Sr. President")
	testSplitCA(t, "de Sant Nicolau (del s. XII; cor gòtic del s. XIV) i de Sant ")
	testSplitCA(t, "Va ser el 5è. classificat.")
	testSplitCA(t, "Va ser el 5è. ", "I l'altre el 4t.")
	testSplitCA(t, "Art. 2.1: Són obligats els...")
	testSplitCA(t, "Arriba fins a les pp. 50-52.")
	testSplitCA(t, "Arriba fins a les pp. XI-XII.")
	testSplitCA(t, "i no ho vol. ", "Malgrat que és així.")
	testSplitCA(t, "i és del vol. 3 de la col·lecció")
	testSplitCA(t, "Els EE. UU. són un país.")
	testSplitCA(t, "Els EE.UU. són un país.")
	testSplitCA(t, "Els ee. uu. són un país.")
	testSplitCA(t, "Els ee.uu. són un país.")
	testSplitCA(t, "Me'n vaig als EE.UU. ", "Bon viatge.")
	testSplitCA(t, "Garcia, Joan (coords.)")
	testSplitCA(t, "fins al curs de 8è. ", "\"No es pot oblidar allò\"")
	testSplitCA(t, "fins al curs de 8è. ", "-No es pot oblidar allò")
	testSplitCA(t, "Aprovació (ca. 2010), suspensió (c. 2011), segle (ca. XIX)")
	testSplitCA(t, "La Dra. Ma. Victòria.")
	testSplitCA(t, "la projectada Sta. Ma. de Gàllecs")
	testSplitCA(t, "El fruit té de 6 a 8 cm de long. i 4 a 6 cm d'ample.")
	testSplitCA(t, "Geiger (Proc. Roy. Soc. 1 de febrer de 1910).")
	testSplitCA(t, "El poble tenia 50 hab. a finals de segle XX.")
	testSplitCA(t, "Vam veure un documental sobre Warner Bros. Cartoons.")
	testSplitCA(t, "Vam veure un documental sobre la Warner Bros. ", "Era boníssim.")
	testSplitCA(t, "La Warner Bros. feia coses que m'agradaven molt.")
	testSplitCA(t, "Introduïu açí el vostre text. ", "o feu servir aquest texts com a a exemple per a alguns errades que LanguageTool hi pot detectat.")

	// Unknown abbreviations inside parentheses
	testSplitCA(t, "(Impren. Disss)")
	testSplitCA(t, "(Impren. 188-disss)")
	testSplitCA(t, "[Impren. Disss]")
	testSplitCA(t, "[Impren. 188-disss]")
	testSplitCA(t, "{Impren. Disss}")
	testSplitCA(t, "{Impren. 188-disss}")
	testSplitCA(t, "(Impren. Disss. Ioo)")
	testSplitCA(t, "(Impren. Disss. Ioo. A. B. Garcia)")
	testSplitCA(t, "Impren. ", "\nDisss")
	testSplitCA(t, "(Impren. ", "\nDisss)")

	// Exception to abbreviations
	testSplitCA(t, "Ell és el número u. ", "Jo el dos.")
	testSplitCA(t, "Té un trau al cap. ", "Cal portar-lo a l'hospital.")
	testSplitCA(t, "Això passa en el PP. ", "Però, per altra banda,")
	testSplitCA(t, "Ceba, all, carabassa, etc. ", "En comprem a la fruiteria.")
	// Units
	testSplitCA(t, "1 500 m/s. ", "Neix a")
	testSplitCA(t, "Són d'1 g. ", "Han estat condicionades.")
	testSplitCA(t, "Són d'1 m. ", "Han estat condicionades.")
	testSplitCA(t, "Hi vivien 50 h. ", "Després el poble va créixer.")
	testSplitCA(t, "L'acte serà a les 15.30 h. de la vesprada.")
	testSplitCA(t, "De 9:00 a 17:00 h. (aproximadament).")
	testSplitCA(t, "Aquesta és la resolució No. 2 de les Corts.")

	// Error: missing space. It is not split in order to trigger other errors.
	testSplitCA(t, "s'hi enfrontà quan G.Oueddei n'esdevingué líder")
	testSplitCA(t, "el jesuïta alemany J.E. Nithard")

	testSplitCA(t, "PERNIL DOLÇ\nBACON\nPEPERONI\nPEBROT VERD\nOLIVES")

	testSplitCA(t, "El framework .NET o ASP.NET o Microsoft.Net")
}
