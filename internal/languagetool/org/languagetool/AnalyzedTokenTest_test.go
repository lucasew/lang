package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.AnalyzedTokenTest

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
