package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInputSentence_Equal(t *testing.T) {
	tok := NewAnalyzedTokenReadings(NewAnalyzedToken("hello", nil, nil))
	a := NewAnalyzedSentence([]*AnalyzedTokenReadings{tok})
	s1 := NewInputSentence(a, "en", "", nil, nil, nil, nil, nil, nil, "ALL", LevelDefault, nil, nil)
	s2 := NewInputSentence(a, "en", "", nil, nil, nil, nil, nil, nil, "ALL", LevelDefault, nil, nil)
	require.True(t, s1.Equal(s2))
	s3 := NewInputSentence(a, "de", "", nil, nil, nil, nil, nil, nil, "ALL", LevelDefault, nil, nil)
	require.False(t, s1.Equal(s3))
	require.NotEmpty(t, s1.String())
}

func TestSimpleInputSentence(t *testing.T) {
	s := NewSimpleInputSentence("hi", "en")
	require.Equal(t, "hi", s.GetText())
	require.True(t, s.Equal(NewSimpleInputSentence("hi", "en")))
}

func TestResultCache(t *testing.T) {
	c := NewResultCache(10)
	key := NewSimpleInputSentence("hi", "en")
	_, ok := c.GetSentenceIfPresent(key)
	require.False(t, ok)
	tok := NewAnalyzedTokenReadings(NewAnalyzedToken("hi", nil, nil))
	sent := NewAnalyzedSentence([]*AnalyzedTokenReadings{tok})
	c.PutSentence(key, sent)
	got, ok := c.GetSentenceIfPresent(key)
	require.True(t, ok)
	require.Equal(t, sent.GetText(), got.GetText())
	require.Greater(t, c.RequestCount(), int64(0))
}
