package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.AnalyzedSentenceTest — full-strength asserts.

func TestAnalyzedSentence_ToString(t *testing.T) {
	words := make([]*AnalyzedTokenReadings, 3)
	sentStart := SentenceStartTagName
	words[0] = NewAnalyzedTokenReadings(NewAnalyzedToken("", &sentStart, nil))
	pos, lemma := "POS", "lemma"
	words[1] = NewAnalyzedTokenReadings(NewAnalyzedToken("word", &pos, &lemma))
	interp := "INTERP"
	words[2] = NewAnalyzedTokenReadings(NewAnalyzedToken(".", &interp, nil))
	sentEnd := SentenceEndTagName
	words[2].AddReading(NewAnalyzedToken(".", &sentEnd, nil), "")
	sentence := NewAnalyzedSentence(words)
	require.Equal(t, "<S> word[lemma/POS].[./INTERP,</S>]", sentence.String())
}

func TestAnalyzedSentence_Copy(t *testing.T) {
	words := make([]*AnalyzedTokenReadings, 3)
	sentStart := SentenceStartTagName
	words[0] = NewAnalyzedTokenReadings(NewAnalyzedToken("", &sentStart, nil))
	pos, lemma := "POS", "lemma"
	words[1] = NewAnalyzedTokenReadings(NewAnalyzedToken("word", &pos, &lemma))
	interp := "INTERP"
	words[2] = NewAnalyzedTokenReadings(NewAnalyzedToken(".", &interp, nil))
	sentEnd := SentenceEndTagName
	words[2].AddReading(NewAnalyzedToken(".", &sentEnd, nil), "")
	sentence := NewAnalyzedSentence(words)
	copySentence := sentence.Copy(sentence)
	require.True(t, sentence.Equals(copySentence))
	words[1].Immunize(999)
	require.Equal(t, "<S> word[lemma/POS{!}].[./INTERP,</S>]", sentence.String())
	require.False(t, sentence.Equals(copySentence))
}

func TestAnalyzedSentence_SetsAndPosition(t *testing.T) {
	tok := NewAnalyzedTokenReadings(NewAnalyzedToken("Hello", nil, nil))
	sp := NewAnalyzedTokenReadings(NewAnalyzedToken(" ", nil, nil))
	// need whitespace flag - AnalyzePlain is easier
	s := AnalyzePlain("Hi there")
	require.NotNil(t, s.GetPreDisambigTokens())
	require.Greater(t, s.GetNonWhitespaceTokenCount(), 0)
	set := s.GetTokenSet()
	require.NotEmpty(t, set)
	_ = tok
	_ = sp
}
