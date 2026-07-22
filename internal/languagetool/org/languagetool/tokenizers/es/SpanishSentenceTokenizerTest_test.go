package es

// Twin of languagetool-language-modules/es/src/test/java/org/languagetool/tokenizers/es/SpanishSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer = new SRXSentenceTokenizer(Spanish.getInstance())
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitES mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitES(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewSpanishSRXSentenceTokenizer()
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of SpanishSentenceTokenizerTest.testTokenize — all non-@Ignore cases, exact equality.
func TestSpanishSentenceTokenizer_Tokenize(t *testing.T) {
	// Simple sentences
	testSplitES(t, "Esto es una frase. ", "Esto es otra frase.")
	testSplitES(t, "Esto es una frase.[34] ", "Esto es otra frase.")
	testSplitES(t, "¿Nos vamos? ", "Hay que irse.")
	testSplitES(t, "¿Vamos? ", "Hay que irse.")
	testSplitES(t, "¡Corre! ", "Hay que irse.")
	testSplitES(t, "1. Artículo primero")

	// Ellipsis
	testSplitES(t, "Entonces... apareció él.")
	testSplitES(t, "Entonces… apareció él.")
	testSplitES(t, "Entoncess… ", "Apareció él.")
	testSplitES(t, "«El tal del tal…» sale de aquí")
	testSplitES(t, "invitarle –cuando sea–… a tomar café.")

	// Initials
	testSplitES(t, "A la atención de A. Comes.")
	testSplitES(t, "A la atenció de À. Comes.")
	testSplitES(t, "Núm. operación 220130000138.")
	testSplitES(t, "N. operación 220130000138.")
	testSplitES(t, "N.º operación 220130000138.")

	// Abbreviations
	testSplitES(t, "las Sras. diputadas")
	testSplitES(t, "No Mr. Spock sino otro.")
	testSplitES(t, "Ver el cap. 24 del libro.")
	testSplitES(t, "Ver el cap. IX del libro.")
	testSplitES(t, "Vive en el núm. 24 de la calle.")
	testSplitES(t, "El Dr. Joan no vendrá.")
	testSplitES(t, "Distingguido Sr. Juan,")
	testSplitES(t, "Muy Hble. Sr. Presidente")
	testSplitES(t, "de San Nicolás (del s. XII; coro gótico del s. XIV) y de San Mateo.")
	testSplitES(t, "fue el 5o. clasificado.")
	testSplitES(t, "Fue el 5º. ", "Y el otro el 4º.")
	testSplitES(t, "Art. 2.1: Estarán obligados...")
	testSplitES(t, "Hasta las pp. 50-52.")
	testSplitES(t, "Hasta las pp. XI-XII.")
	testSplitES(t, "y es del vol. 3 de la colección")
	testSplitES(t, "En EE.UU.")
	testSplitES(t, "En EE. UU. por los DD. HH. después de los JJ. OO.")
	testSplitES(t, "En U.S.A. años 30.")
	testSplitES(t, "En U. S. A. años 30.")
	testSplitES(t, "P. ej. esto.")
	testSplitES(t, "Ahora p. ej. esto.")
	testSplitES(t, "Ahora p. e. esto.")
	testSplitES(t, "Son las 5hrs. del domingo.")
	testSplitES(t, "Son las 2as. jornadas.")
	testSplitES(t, "En EE.UU. esto no pasa.")
	testSplitES(t, "En EE. UU. esto no pasa.")
	testSplitES(t, "Me voy a EE. UU. ", "Buen viaje.")
	testSplitES(t, "Uno (ca. 2010), dos (c. 2011), tres (ca. XIX), cuatro (c. XX)")
	testSplitES(t, "Ayto. de Madrid.")
	testSplitES(t, "¿Quién sabe hablar francés mejor: Tom o Mary?")
	testSplitES(t, "Hola, Albert: ", "Me puedes decir tu correo?")
	testSplitES(t, "LanguageTooler GmbH recaudará de tu cuenta a través de GoCardless Ltd. la cantidad debajo mencionada.")
	testSplitES(t, "El fruto es una nuez de 6 a 8 cm de long. y 4 a 6 cm de ancho")
	testSplitES(t, "Geiger (Proc. Roy. Soc. 1 de febrero de 1910).")
	testSplitES(t, "Es la resolución No. 2 del parlamento,")
	testSplitES(t, "Con un dto. del 50 %")
	testSplitES(t, "DTO. 50%")
	testSplitES(t, "DTO. DEL 50%")
	testSplitES(t, "Ayto. del Ferrol")
	testSplitES(t, "En el ayto. del municipio.")
	testSplitES(t, "Compré 6 Ltrs. de leche.")
	testSplitES(t, "Mi profesora trabaja los lun., mié. y vie.; los juev. y los dom., no.")

	// Exception to abbreviations
	testSplitES(t, "Esto pasa el PP. ", "Pero, por otra parte,")
	testSplitES(t, "Cebolla, ajo, calabaza, etc. ", "Compramos en fruitería.")
	// Units
	testSplitES(t, "1 500 m/s. ", "Nacen en")
	testSplitES(t, "Son de 1 g. ", "Han sido acondicionadas.")
	testSplitES(t, "Son de 1 m. ", "Han sido acondicionadas.")
	testSplitES(t, "Vivían 50 h. ", "Después el pueblo creció.")
	testSplitES(t, "El acto será a las 15.30 h. de la tarde.")
	testSplitES(t, "Se calcula un Vol. aproximado de 3.5 ml.")
	testSplitES(t, "Se aplicaron 5 cc. de anestesia local durante el procedimiento.")
	testSplitES(t, "El dispositivo opera a 2400 MHz. lo que garantiza una conexión estable.")
	testSplitES(t, "El paciente fue evaluado tras 12 hr. de observación continua.")
	testSplitES(t, "El medicamento se administró pb. dos veces al día.")
	testSplitES(t, "La reunión se realizará en la 3a. sala del edificio principal.")
	testSplitES(t, "El informe médico indica Dx. confirmado de neumonía.")
	testSplitES(t, "Cursó la lic. en Administración de Empresas.")
	testSplitES(t, "Cursó la Lic. En Administración de Empresas.")

	// Error: missing space. It is not split in order to trigger other errors.
	testSplitES(t, "cuando G.Oueddei se convierte en líder")
	testSplitES(t, "El jesuita alemán J.E. Nithard")
}
