package crh

// Twin of languagetool-language-modules/crh/src/test/java/org/languagetool/tokenizers/crh/CrimeanTatarSRXSentenceTokenizerTest.java
// Java: stokenizer = new SRXSentenceTokenizer(LANGUAGE) // CrimeanTatar short code "crh"
// @Before setUp: stokenizer.setSingleLineBreaksMarksParagraph(true)
// Java: testSplit → TestTools.testSplit(sentences, stokenizer) — join parts, tokenize, expect same parts.
// Commented TODO case is not ported as active.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitCRH mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitCRH(t *testing.T, stokenizer *CrimeanTatarSRXSentenceTokenizer, parts ...string) {
	t.Helper()
	joined := strings.Join(parts, "")
	got := stokenizer.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of CrimeanTatarSRXSentenceTokenizerTest.testTokenize — 2 active cases only, exact equality.
func TestCrimeanTatarSRXSentenceTokenizer_Tokenize(t *testing.T) {
	// private final SentenceTokenizer stokenizer = new SRXSentenceTokenizer(LANGUAGE);
	// accept \n as paragraph:
	stokenizer := NewCrimeanTatarSRXSentenceTokenizer()
	// @Before setUp:
	stokenizer.SetSingleLineBreaksMarksParagraph(true)

	// NOTE: sentences here need to end with a space character so they
	// have correct whitespace when appended:
	testSplitCRH(t, stokenizer, "Yapraqlar töküldi. ")
	testSplitCRH(t, stokenizer, "Yapraqlar töküldi. ", "Otlar-ölenler sarardı, soldılar.")
	//TODO: not ported as active — Java has it commented out:
	// testSplit("– Afu etiñiz, ocam! – dedim men, – Selâmet mıtlaqa ketmek kerekmi?");
}
