package ja

// Twin of languagetool-language-modules/ja/src/test/java/org/languagetool/tokenizers/ja/JapaneseSRXSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer = new SRXSentenceTokenizer(new Japanese())  // short code "ja"
// No setSingleLineBreaksMarksParagraph — use SRX defaults for Japanese.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitJA mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitJA(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewJapaneseSRXSentenceTokenizer()
	// default paragraph mode — do NOT invent flags unless Java sets them
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of JapaneseSRXSentenceTokenizerTest.testTokenize — all active cases, exact equality.
func TestJapaneseSRXSentenceTokenizer_Tokenize(t *testing.T) {
	testSplitJA(t, "これはテスト用の文です。")
	testSplitJA(t, "これはテスト用の文です。", "追加のテスト用の文です。")
	testSplitJA(t, "これは、テスト用の文です。")
	testSplitJA(t, "テスト用の文です！", "追加のテスト用の文です。")
	testSplitJA(t, "テスト用の文です... ", "追加のテスト用の文です。")
	testSplitJA(t, "アドレスはhttp://www.test.deです。")

	testSplitJA(t, "これは(!)の文です。")
	testSplitJA(t, "これは(!!)の文です。")
	testSplitJA(t, "これは(?)の文です。")
	testSplitJA(t, "これは(??)の文です。")
}
