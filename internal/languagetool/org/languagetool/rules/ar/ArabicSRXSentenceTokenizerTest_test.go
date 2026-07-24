package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicSRXSentenceTokenizerTest.java
// Java package is org.languagetool.rules.ar; class ArabicSRXSentenceTokenizerTest.
// Java: stokenizer = new SRXSentenceTokenizer(new Arabic())  // short code "ar"
// Java: testSplit → TestTools.testSplit(sentences, stokenizer) — join parts, tokenize, expect same parts.
// No setSingleLineBreaksMarksParagraph — use SRX defaults for Arabic.

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// testSplitAR mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitAR(t *testing.T, stokenizer *tokenizers.SRXSentenceTokenizer, parts ...string) {
	t.Helper()
	joined := strings.Join(parts, "")
	got := stokenizer.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of ArabicSRXSentenceTokenizerTest.test — all 4 active cases, exact equality.
func TestArabicSRXSentenceTokenizer_Test(t *testing.T) {
	// private final SRXSentenceTokenizer stokenizer = new SRXSentenceTokenizer(new Arabic());
	// default paragraph mode — do NOT invent flags unless Java sets them
	stokenizer := tokenizers.NewSRXSentenceTokenizer("ar")

	testSplitAR(t, stokenizer, "مشوار التعلم طويل.")
	testSplitAR(t, stokenizer, "هل ستنام الليلة؟")
	testSplitAR(t, stokenizer, "قُل: توْأمٌ، وتوْأمانِ: وقلْ: هذانِ توْأمانِ.. ")
	testSplitAR(t, stokenizer, "قلْ: هذِهِ توْأمُ «هذا»، (وقلْ: هذِهِ توْأمةُ هذا)، وقلْ: هذانِ توْأمٌ!")
}
