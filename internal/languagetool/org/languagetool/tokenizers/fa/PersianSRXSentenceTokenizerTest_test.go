package fa

// Twin of languagetool-language-modules/fa/src/test/java/org/languagetool/tokenizers/PersianSRXSentenceTokenizerTest.java
// Java package is org.languagetool.tokenizers (not fa); class PersianSRXSentenceTokenizerTest.
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer = new SRXSentenceTokenizer(new Persian())  // short code "fa"
// No setSingleLineBreaksMarksParagraph — use SRX defaults for Persian.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitFA mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitFA(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewPersianSRXSentenceTokenizer()
	// default paragraph mode — do NOT invent flags unless Java sets them
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of PersianSRXSentenceTokenizerTest.test — all active cases, exact equality.
func TestPersianSRXSentenceTokenizer_Tokenize(t *testing.T) {
	// NOTE: sentences here need to end with a space character so they
	// have correct whitespace when appended:
	testSplitFA(t, "این یک جمله است. ", "جملهٔ بعدی")
	testSplitFA(t, "آیا این یک جمله است؟ ", "جملهٔ بعدی")
	testSplitFA(t, "یک جمله!!! ", "جملهٔ بعدی")

	testSplitFA(t, "جملهٔ اول... خوب نیست؟ ", "جملهٔ دوم.")
	testSplitFA(t, "جملهٔ اول (...) ادامهٔ متن. ")
	testSplitFA(t, "جملهٔ اول [...] ادامهٔ متن. ")
}
