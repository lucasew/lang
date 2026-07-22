package ml

// Twin of languagetool-language-modules/ml/src/test/java/org/languagetool/tokenizers/ml/MalayalamSRXSentenceTokenizerTest.java
// Java: stokenizer = new SRXSentenceTokenizer(new Malayalam()) // short code "ml"
// No setSingleLineBreaksMarksParagraph — use SRX defaults only.
// Java: testSplit → TestTools.testSplit(sentences, stokenizer) — join parts, tokenize, expect same parts.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitML mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitML(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewMalayalamSRXSentenceTokenizer()
	// default paragraph mode — do NOT invent flags unless Java sets them
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of MalayalamSRXSentenceTokenizerTest.testTokenize — single active multi-part case, exact equality.
func TestMalayalamSRXSentenceTokenizer_Tokenize(t *testing.T) {
	// NOTE: first sentence ends with a space so whitespace is correct when appended.
	// Preserve exact Unicode from Java source (includes ZWNJ U+200C in സോഫ്റ്റ്‌വെയർ forms).
	testSplitML(t,
		"1983 ൽ റിച്ചാർഡ് സ്റ്റാൾമാൻ സ്ഥാപിച്ച ഗ്നു എന്ന സംഘടനയിൽ നിന്നും വളർന്നു വന്ന സോഫ്റ്റ്‌വെയറും ടൂളുകളുമാണ് ഇന്ന് ഗ്നൂ/ലിനക്സിൽ ലഭ്യമായിട്ടുള്ള സോഫ്റ്റ്‌വെയറിൽ സിംഹഭാഗവും. ",
		"ഗ്നു സംഘത്തിന്റെ മുഖ്യലക്ഷ്യം സ്വതന്ത്ര സോഫ്റ്റ്‌വെയറുകൾ മാത്രം ഉപയോഗിച്ചുകൊണ്ട് യുണിക്സ് പോലുള്ള ഒരു ഓപ്പറേറ്റിംഗ് സിസ്റ്റം നിർമ്മിക്കുന്നതായിരുന്നു.",
	)
}
