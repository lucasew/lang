package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.AnalyzedTokenTest — full-strength asserts (1:1).

func TestAnalyzedToken_ToString(t *testing.T) {
	pos := "POS"
	lemma := "lemma"
	testToken := NewAnalyzedToken("word", &pos, &lemma)
	require.Equal(t, "lemma/POS", testToken.String())
	require.NotNil(t, testToken.GetLemma())
	require.Equal(t, "lemma", *testToken.GetLemma())

	testToken2 := NewAnalyzedToken("word", &pos, nil)
	require.Equal(t, "word/POS", testToken2.String())
	require.Nil(t, testToken2.GetLemma())
	require.Equal(t, "word", testToken2.GetToken())
}

func TestAnalyzedToken_Matches(t *testing.T) {
	pos := "POS"
	lemma := "lemma"
	testToken1 := NewAnalyzedToken("word", &pos, &lemma)

	require.False(t, testToken1.Matches(NewAnalyzedToken("", nil, nil)))
	require.True(t, testToken1.Matches(NewAnalyzedToken("word", nil, nil)))
	require.True(t, testToken1.Matches(NewAnalyzedToken("word", &pos, nil)))
	require.True(t, testToken1.Matches(NewAnalyzedToken("word", &pos, &lemma)))

	pos1 := "POS1"
	require.False(t, testToken1.Matches(NewAnalyzedToken("word", &pos1, &lemma)))
	require.False(t, testToken1.Matches(NewAnalyzedToken("word1", &pos, &lemma)))
	lemma1 := "lemma1"
	require.False(t, testToken1.Matches(NewAnalyzedToken("word", &pos, &lemma1)))
	require.True(t, testToken1.Matches(NewAnalyzedToken("", &pos, &lemma)))
	require.True(t, testToken1.Matches(NewAnalyzedToken("", nil, &lemma)))
}

// Extra behavior from AnalyzedToken.java (not covered by AnalyzedTokenTest.java).

func TestAnalyzedToken_HasNoTag_AndWhitespace(t *testing.T) {
	// null POS → hasNoTag
	tok := NewAnalyzedToken("x", nil, nil)
	require.True(t, tok.HasNoTag())
	require.Equal(t, "x/null", tok.String()) // Java string concat of null → "null"

	// real POS → has tag
	pos := "NN"
	tok2 := NewAnalyzedToken("x", &pos, nil)
	require.False(t, tok2.HasNoTag())

	// SENT_END / PARA_END (untrimmed param) → hasNoTag
	sent := SentenceEndTagName
	tok3 := NewAnalyzedToken(".", &sent, nil)
	require.True(t, tok3.HasNoTag())
	para := ParagraphEndTagName
	tok4 := NewAnalyzedToken("\n", &para, nil)
	require.True(t, tok4.HasNoTag())

	// Java: hasNoPOSTag uses original posTag *before* trim — padded special tag is NOT no-tag
	padded := " " + SentenceEndTagName + " "
	tok5 := NewAnalyzedToken(".", &padded, nil)
	require.Equal(t, SentenceEndTagName, *tok5.GetPOSTag()) // stored trimmed
	require.False(t, tok5.HasNoTag())                      // original param != SENT_END

	// posTag is trimmed on store
	spaced := "  VB  "
	tok6 := NewAnalyzedToken("run", &spaced, nil)
	require.Equal(t, "VB", *tok6.GetPOSTag())

	// setNoPOSTag override
	tok6.SetNoPOSTag(true)
	require.True(t, tok6.HasNoTag())

	// whitespaceBefore
	require.False(t, tok6.IsWhitespaceBefore())
	tok6.SetWhitespaceBefore(true)
	require.True(t, tok6.IsWhitespaceBefore())
}

func TestAnalyzedToken_Equals(t *testing.T) {
	pos := "POS"
	lemma := "lemma"
	a := NewAnalyzedToken("word", &pos, &lemma)
	b := NewAnalyzedToken("word", &pos, &lemma)
	require.True(t, a.Equals(b))
	require.True(t, a.Equals(a))
	require.False(t, a.Equals(nil))

	a.SetWhitespaceBefore(true)
	require.False(t, a.Equals(b))
	b.SetWhitespaceBefore(true)
	require.True(t, a.Equals(b))

	otherPos := "XX"
	c := NewAnalyzedToken("word", &otherPos, &lemma)
	c.SetWhitespaceBefore(true)
	require.False(t, a.Equals(c))
}
