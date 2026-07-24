package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/tokenizers/pt/PortugueseSRXSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// tokenizer = new SRXSentenceTokenizer(Portuguese.getInstance())  // short code "pt"
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitPT mirrors Java private testSplit → TestTools.testSplit(sentences, tokenizer).
func testSplitPT(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewPortugueseSRXSentenceTokenizer()
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of PortugueseSRXSentenceTokenizerTest.testTokenize — all non-@Ignore cases, exact equality.
func TestPortugueseSRXSentenceTokenizer_Tokenize(t *testing.T) {
	// NOTE: sentences here need to end with a space character so they
	// have correct whitespace when appended:
	testSplitPT(t, "Cola o teu próprio texto aqui.")
	testSplitPT(t, "Cola o teu próprio texto aqui. ", "Ou verifica este texto.")

	// Missing white space between sentences: do not split
	testSplitPT(t, "Esta é a primeira frase.Esta é a segunda.")

	// Basic sentence splitting
	testSplitPT(t, "O Brasil é um país muito grande. ", "Tem muitos estados e cidades.")
	testSplitPT(t, "Hoje está fazendo muito calor. ", "Vamos tomar sorvete. ", "Que boa ideia!")
	testSplitPT(t, "Você gosta de futebol? ", "Eu adoro!")

	// Abbreviations that should NOT split
	testSplitPT(t, "O Sr. João foi ao mercado.")
	testSplitPT(t, "A Sra. Silva mora na Rua das Flores.")
	testSplitPT(t, "O Dr. Carlos atendeu o paciente ontem.")
	testSplitPT(t, "A Dra. Ana é especialista em pediatria.")
	testSplitPT(t, "O Prof. Souza deu uma aula excelente.")
	testSplitPT(t, "Comprei frutas, legumes, etc. no supermercado.")
	testSplitPT(t, "São precisos documentos, certidões, etc. para o processo.")
	testSplitPT(t, "Havia problemas de logística, infraestrutura, etc. ", "Tudo precisava ser resolvido.")
	testSplitPT(t, "Comprei maçãs, peras, laranjas, etc. ", "Depois fui para casa.")
	testSplitPT(t, "O endereço é Av. Paulista, 1000.")
	testSplitPT(t, "Moro na R. das Flores, n.º 25.")
	testSplitPT(t, "Consulte o cap. 3 para mais informações.")
	testSplitPT(t, "Veja a fig. 2 abaixo.")
	testSplitPT(t, "O evento ocorreu em jan. de 2023.")

	// Abbreviations followed by sentence boundary
	testSplitPT(t, "O contrato foi assinado ontem. ", "Depois foi registrado em cartório.")
	testSplitPT(t, "O Prof. Silva chegou tarde. ", "A aula começou com atraso.")
	testSplitPT(t, "Consulte o Dr. Almeida. ", "Ele poderá ajudá-lo.")

	// Question marks and exclamation marks
	testSplitPT(t, "Você viu o filme? ", "Eu achei incrível!")
	testSplitPT(t, "Como você está? ", "Estou bem, obrigado.")
	testSplitPT(t, "Que dia bonito! ", "Vamos passear no parque.")
	testSplitPT(t, "Será que vai chover? ", "Melhor levar guarda-chuva.")

	// Ellipsis
	testSplitPT(t, "Não sei o que dizer... ", "É uma situação muito difícil.")
	testSplitPT(t, "Ele hesitou por um momento... e então decidiu partir.")
	testSplitPT(t, "Ele viria ... ?")
	testSplitPT(t, "Ele viria, ... ?")

	// Ordinal numbers with dot
	testSplitPT(t, "O 1.º lugar foi do Brasil.")
	testSplitPT(t, "A 2.ª colocada foi a Argentina. ", "O 3.º lugar ficou com o Uruguai.")

	// Numbers with dots (should not split)
	testSplitPT(t, "O evento começa às 10.30 e termina às 12.00.")
	testSplitPT(t, "O texto tem 3.500 palavras ao todo.")

	// Initials and proper names
	testSplitPT(t, "J. K. Rowling é a autora de Harry Potter.")
	testSplitPT(t, "O presidente L. I. Lula assinou o decreto. ", "Será implementado em breve.")

	// Quotes with sentence boundaries
	testSplitPT(t, "\"Vou embora!\", avisou ela. ", "Todos ficaram tristes.")
	testSplitPT(t, "\"Não aguento mais!\", gritou ela. ", "Todos olharam.")

	// URLs (should not split)
	testSplitPT(t, "Acesse o site em http://www.exemplo.com.br para mais informações.")

	// Mixed punctuation
	testSplitPT(t, "O Brasil ganhou! ", "Que festa incrível! ", "Todos comemoraram.")
}
