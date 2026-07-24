package lt

// Twin of languagetool-language-modules/lt/src/test/java/org/languagetool/tokenizers/lt/LithuanianSRXSentenceTokenizerTest.java
// Java: stokenizer = new SRXSentenceTokenizer(new Lithuanian()) // short code "lt"
// No setSingleLineBreaksMarksParagraph — use SRX defaults only.
// Java: testSplit → TestTools.testSplit(sentences, stokenizer) — join parts, tokenize, expect same parts.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitLT mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitLT(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewLithuanianSRXSentenceTokenizer()
	// default paragraph mode — do NOT invent flags unless Java sets them
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of LithuanianSRXSentenceTokenizerTest.testTokenize — single active multi-part case, exact equality.
func TestLithuanianSRXSentenceTokenizer_Tokenize(t *testing.T) {
	// NOTE: first sentence ends with a space so whitespace is correct when appended.
	// Preserve exact Unicode: en dash (–) and Lithuanian diacritics (ė, ų, į, etc.).
	testSplitLT(t,
		"Linux – laisvos operacinės sistemos branduolio (kernel) pavadinimas. ",
		"Dažnai taip sutrumpintai vadinama ir bendrai visa Unix-tipo operacinė sistema naudojanti Linux branduolį kartu su sisteminėmis programomis bei bibliotekomis iš GNU projekto.",
	)
}
