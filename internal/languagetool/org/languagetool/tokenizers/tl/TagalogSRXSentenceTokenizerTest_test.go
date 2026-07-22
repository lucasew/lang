package tl

// Twin of languagetool-language-modules/tl/src/test/java/org/languagetool/tokenizers/tl/TagalogSRXSentenceTokenizerTest.java
// Java: stokenizer = new SRXSentenceTokenizer(new Tagalog()) // short code "tl"
// No setSingleLineBreaksMarksParagraph — use SRX defaults only.
// Java: testSplit → TestTools.testSplit(sentences, stokenizer) — join parts, tokenize, expect same parts.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitTL mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitTL(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewTagalogSRXSentenceTokenizer()
	// default paragraph mode — do NOT invent flags unless Java sets them
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of TagalogSRXSentenceTokenizerTest.testTokenize — single active multi-part case, exact equality.
// Second part is one string equal to Java's concatenated second argument.
func TestTagalogSRXSentenceTokenizer_Tokenize(t *testing.T) {
	// NOTE: first sentence ends with a space so whitespace is correct when appended.
	testSplitTL(t,
		"Ang Linux ay isang operating system kernel para sa mga operating system na humahalintulad sa Unix. ",
		"Isa ang Linux sa mga pinaka-prominanteng halimbawa ng malayang software at pagsasagawa ng open source; "+
			"madalas, malayang mapapalitan, gamitin, at maipamahagi ninuman ang "+
			"lahat ng pinag-ugatang source code (pinagmulang kodigo).",
	)
}
