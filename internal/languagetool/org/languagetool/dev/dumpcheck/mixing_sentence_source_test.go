package dumpcheck

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMixingSentenceSource_Alternates(t *testing.T) {
	// Two plain sources with long-enough lines
	a := NewPlainTextSentenceSource(strings.NewReader(
		"Alpha sentence number one is long.\nAlpha sentence number two is long.\n"))
	b := NewPlainTextSentenceSource(strings.NewReader(
		"Beta sentence number one is long.\nBeta sentence number two is long.\n"))
	// Disable source name being empty - PlainText uses currentURL which may be empty
	mix := NewMixingSentenceSource([]SentenceSource{a, b})
	require.True(t, mix.HasNext())

	var texts []string
	for mix.HasNext() {
		s, err := mix.Next()
		require.NoError(t, err)
		texts = append(texts, s.GetText())
	}
	require.Len(t, texts, 4)
	// alternate: a, b, a, b
	require.True(t, strings.HasPrefix(texts[0], "Alpha"))
	require.True(t, strings.HasPrefix(texts[1], "Beta"))
	require.True(t, strings.HasPrefix(texts[2], "Alpha"))
	require.True(t, strings.HasPrefix(texts[3], "Beta"))
}

func TestMixingSentenceSource_ExhaustedSourceRemoved(t *testing.T) {
	a := NewPlainTextSentenceSource(strings.NewReader("Only alpha sentence is quite long.\n"))
	b := NewPlainTextSentenceSource(strings.NewReader(
		"Beta sentence number one is long.\nBeta sentence number two is long.\n"))
	mix := NewMixingSentenceSource([]SentenceSource{a, b})
	var texts []string
	for mix.HasNext() {
		s, err := mix.Next()
		require.NoError(t, err)
		texts = append(texts, s.GetText())
	}
	require.Len(t, texts, 3)
	require.True(t, strings.HasPrefix(texts[0], "Only"))
	require.True(t, strings.HasPrefix(texts[1], "Beta"))
	require.True(t, strings.HasPrefix(texts[2], "Beta"))
}

func TestMixingSentenceSource_TatoebaAndPlain(t *testing.T) {
	tat := NewTatoebaSentenceSource(strings.NewReader(
		"1\teng\tTatoeba sample sentence is long enough here.\n"))
	plain := NewPlainTextSentenceSource(strings.NewReader(
		"Plain sample sentence is long enough here.\n"))
	mix := NewMixingSentenceSource([]SentenceSource{tat, plain})
	s1, err := mix.Next()
	require.NoError(t, err)
	require.Equal(t, "tatoeba", s1.GetSource())
	s2, err := mix.Next()
	require.NoError(t, err)
	// plain source name is empty until # source: line
	require.Equal(t, "Tatoeba sample sentence is long enough here.", s1.GetText())
	require.Equal(t, "Plain sample sentence is long enough here.", s2.GetText())
	dist := mix.GetSourceDistribution()
	require.Equal(t, 1, dist["tatoeba"])
}
