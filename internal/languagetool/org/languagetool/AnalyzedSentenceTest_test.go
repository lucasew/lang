package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.AnalyzedSentenceTest — full asserts; impl stubbed.

func TestAnalyzedSentence_ToString(t *testing.T) {
	words := make([]*AnalyzedTokenReadings, 3)
	sentStart := "SENT_START"
	words[0] = NewAnalyzedTokenReadings(NewAnalyzedToken("", &sentStart, nil))
	pos, lemma := "POS", "lemma"
	words[1] = NewAnalyzedTokenReadings(NewAnalyzedToken("word", &pos, &lemma))
	interp := "INTERP"
	words[2] = NewAnalyzedTokenReadings(NewAnalyzedToken(".", &interp, nil))
	sentEnd := "SENT_END"
	words[2].AddReading(NewAnalyzedToken(".", &sentEnd, nil), "")
	sentence := NewAnalyzedSentence(words)
	require.Equal(t, "<S> word[lemma/POS].[./INTERP,</S>]", sentence.String())
}

func TestAnalyzedSentence_Copy(t *testing.T) {
	words := make([]*AnalyzedTokenReadings, 3)
	sentStart := "SENT_START"
	words[0] = NewAnalyzedTokenReadings(NewAnalyzedToken("", &sentStart, nil))
	pos, lemma := "POS", "lemma"
	words[1] = NewAnalyzedTokenReadings(NewAnalyzedToken("word", &pos, &lemma))
	interp := "INTERP"
	words[2] = NewAnalyzedTokenReadings(NewAnalyzedToken(".", &interp, nil))
	sentEnd := "SENT_END"
	words[2].AddReading(NewAnalyzedToken(".", &sentEnd, nil), "")
	sentence := NewAnalyzedSentence(words)
	copySentence := sentence.Copy(sentence)
	require.True(t, sentence.Equals(copySentence))
	words[1].Immunize(999)
	require.Equal(t, "<S> word[lemma/POS{!}].[./INTERP,</S>]", sentence.String())
	require.False(t, sentence.Equals(copySentence))
}
