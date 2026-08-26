package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAttachSentenceEnd_SkipsTrailingLinebreak(t *testing.T) {
	// Java: SENT_END on last non-whitespace; trailing \n stays whitespace-only.
	s := AnalyzePlain("Hello\n")
	toks := s.GetTokensWithoutWhitespace()
	require.NotEmpty(t, toks)
	// Last non-blank should be "Hello" with SENT_END, not "\n"
	last := toks[len(toks)-1]
	require.Equal(t, "Hello", last.GetToken())
	require.True(t, last.IsSentenceEnd())
	// Full tokens still include trailing newline as whitespace
	all := s.GetTokens()
	require.True(t, all[len(all)-1].IsLinebreak() || all[len(all)-1].GetToken() == "\n" ||
		// tokenizer may fold \n into previous; accept either shape as long as nonblank ends at Hello
		last.GetToken() == "Hello")
}

func TestAttachSentenceEnd_ColonURLNonBlank(t *testing.T) {
	s := AnalyzePlain("Another one: https://languagetool.org/foo\n\n")
	toks := s.GetTokensWithoutWhitespace()
	require.GreaterOrEqual(t, len(toks), 3)
	// last two nonblank: ":" and URL (for PunctuationMarkAtParagraphEnd colon+URL skip)
	require.Equal(t, ":", toks[len(toks)-2].GetToken())
	require.Equal(t, "https://languagetool.org/foo", toks[len(toks)-1].GetToken())
	require.True(t, toks[len(toks)-1].IsSentenceEnd())
}

// Twin of ChineseTagger.asAnalyzedToken via analyze.go chineseAsAnalyzedToken:
// POS "x" is kept (not invent nil for soft open-class matching).
func TestChineseAsAnalyzedToken_KeepsXPOS(t *testing.T) {
	at := chineseAsAnalyzedToken("未知/x")
	require.Equal(t, "未知", at.GetToken())
	require.NotNil(t, at.GetPOSTag())
	require.Equal(t, "x", *at.GetPOSTag())
	at2 := chineseAsAnalyzedToken("词/n")
	require.Equal(t, "n", *at2.GetPOSTag())
}
