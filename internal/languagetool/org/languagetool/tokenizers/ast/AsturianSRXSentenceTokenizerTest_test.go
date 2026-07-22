package ast

// Twin of languagetool-language-modules/ast/src/test/java/org/languagetool/tokenizers/ast/AsturianSRXSentenceTokenizerTest.java
// Java: one SRXSentenceTokenizer(new Asturian()) field; testSplit → TestTools.testSplit(sentences, stokenizer).
// Default ctor: setSingleLineBreaksMarksParagraph(false). Java mutates the same instance mid-test.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitAST mirrors Java private testSplit on a shared stokenizer (exact join/parts equality).
func testSplitAST(t *testing.T, tok *AsturianSRXSentenceTokenizer, parts ...string) {
	t.Helper()
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of AsturianSRXSentenceTokenizerTest.testTokenize — all 5 active cases, exact equality.
func TestAsturianSRXSentenceTokenizer_Tokenize(t *testing.T) {
	// private final SRXSentenceTokenizer stokenizer = new SRXSentenceTokenizer(new Asturian());
	// default: setSingleLineBreaksMarksParagraph(false) in SRXSentenceTokenizer ctor
	stokenizer := NewAsturianSRXSentenceTokenizer()

	testSplitAST(t, stokenizer,
		"De secute, los hackers de Minix aportaron idegues y códigu al núcleu Linux, y güey recibiera contribuciones de miles de programadores. ",
		"Torvalds sigue lliberando nueves versiones del núcleu, consolidando aportes d'otros programadores y faciendo cambios el mesmu.")

	stokenizer.SetSingleLineBreaksMarksParagraph(false)
	testSplitAST(t, stokenizer, "De secute,\nlos hackers de Minix...")
	testSplitAST(t, stokenizer, "De secute,\n\n", "los hackers de Minix...")

	stokenizer.SetSingleLineBreaksMarksParagraph(true)
	testSplitAST(t, stokenizer, "De secute,\n", "los hackers de Minix...")
	testSplitAST(t, stokenizer, "De secute,\n", "\n", "los hackers de Minix...")
}
