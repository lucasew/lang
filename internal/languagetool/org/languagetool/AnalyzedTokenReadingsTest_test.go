package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.AnalyzedTokenReadingsTest — full-strength asserts.

func TestAnalyzedTokenReadings_NewTags(t *testing.T) {
	pos := "POS"
	lemma := "lemma"
	tokenReadings := NewAnalyzedTokenReadings(NewAnalyzedToken("word", &pos, &lemma))
	require.Equal(t, false, tokenReadings.IsLinebreak())
	require.Equal(t, false, tokenReadings.IsSentenceEnd())
	require.Equal(t, false, tokenReadings.IsParagraphEnd())
	require.Equal(t, false, tokenReadings.IsSentenceStart())
	tokenReadings.SetSentEnd()
	require.Equal(t, false, tokenReadings.IsSentenceStart())
	require.Equal(t, true, tokenReadings.IsSentenceEnd())

	tokenReadings = NewAnalyzedTokenReadings(NewAnalyzedToken("word", nil, &lemma))
	sentEnd := SentenceEndTagName
	tokenReadings.AddReading(NewAnalyzedToken("word", &sentEnd, nil), "")
	require.Equal(t, true, tokenReadings.IsSentenceEnd())
	require.Equal(t, false, tokenReadings.IsParagraphEnd())
	paraEnd := ParagraphEndTagName
	tokenReadings.AddReading(NewAnalyzedToken("word", &paraEnd, nil), "")
	require.Equal(t, true, tokenReadings.IsParagraphEnd())
	require.Equal(t, false, tokenReadings.IsSentenceStart())
	sentStart := SentenceStartTagName
	tokenReadings.AddReading(NewAnalyzedToken("word", &sentStart, nil), "")
	require.Equal(t, false, tokenReadings.IsSentenceStart())

	aTok := NewAnalyzedToken("word", &pos, &lemma)
	aTok.SetWhitespaceBefore(true)
	tokenReadings = NewAnalyzedTokenReadings(aTok)
	require.True(t, aTok.Equals(tokenReadings.GetAnalyzedToken(0)))
	aTok2 := NewAnalyzedToken("word", &pos, &lemma)
	require.True(t, !aTok2.Equals(tokenReadings.GetAnalyzedToken(0)))
	aTok3 := NewAnalyzedToken("word", &pos, &lemma)
	aTok3.SetWhitespaceBefore(true)
	require.True(t, aTok3.Equals(tokenReadings.GetAnalyzedToken(0)))

	testReadings := NewAnalyzedTokenReadings(aTok3)
	testReadings.RemoveReading(aTok3, "")
	require.True(t, testReadings.GetReadingsLength() == 1)
	require.Equal(t, "word", testReadings.GetToken())
	require.True(t, !testReadings.HasPosTag("POS"))

	testReadings.LeaveReading(aTok2)
	require.Equal(t, "word", testReadings.GetToken())
	require.True(t, !testReadings.HasPosTag("POS"))

	testReadings.RemoveReading(aTok2, "")
	require.Equal(t, "word", testReadings.GetToken())
	require.True(t, !testReadings.HasPosTag("POS"))
}

func TestAnalyzedTokenReadings_ToString(t *testing.T) {
	pos := "POS"
	lemma := "lemma"
	tokenReadings := NewAnalyzedTokenReadings(NewAnalyzedToken("word", &pos, &lemma))
	require.Equal(t, "word[lemma/POS*]", tokenReadings.String())
	pos2, lemma2 := "POS2", "lemma2"
	aTok2 := NewAnalyzedToken("word", &pos2, &lemma2)
	tokenReadings.AddReading(aTok2, "")
	require.Equal(t, "word[lemma/POS*,lemma2/POS2*]", tokenReadings.String())
}

func TestAnalyzedTokenReadings_HasPosTag(t *testing.T) {
	pos := "POS:FOO:BAR"
	lemma := "lemma"
	tokenReadings := NewAnalyzedTokenReadings(NewAnalyzedToken("word", &pos, &lemma))
	require.True(t, tokenReadings.HasPosTag("POS:FOO:BAR"))
	require.False(t, tokenReadings.HasPosTag("POS:FOO:bar"))
	require.False(t, tokenReadings.HasPosTag("POS:FOO"))
	require.False(t, tokenReadings.HasPosTag("xaz"))
}

func TestAnalyzedTokenReadings_HasPartialPosTag(t *testing.T) {
	pos := "POS:FOO:BAR"
	lemma := "lemma"
	tokenReadings := NewAnalyzedTokenReadings(NewAnalyzedToken("word", &pos, &lemma))
	require.True(t, tokenReadings.HasPartialPosTag("POS:FOO:BAR"))
	require.True(t, tokenReadings.HasPartialPosTag("POS:FOO:"))
	require.True(t, tokenReadings.HasPartialPosTag("POS:FOO"))
	require.True(t, tokenReadings.HasPartialPosTag(":FOO:"))
	require.True(t, tokenReadings.HasPartialPosTag("FOO:BAR"))
	require.False(t, tokenReadings.HasPartialPosTag("POS:FOO:BARX"))
	require.False(t, tokenReadings.HasPartialPosTag("POS:foo:BAR"))
	require.False(t, tokenReadings.HasPartialPosTag("xaz"))
}

func TestAnalyzedTokenReadings_MatchesPosTagRegex(t *testing.T) {
	pos := "POS:FOO:BAR"
	lemma := "lemma"
	tokenReadings := NewAnalyzedTokenReadings(NewAnalyzedToken("word", &pos, &lemma))
	require.True(t, tokenReadings.MatchesPosTagRegex("POS:FOO:BAR"))
	require.True(t, tokenReadings.MatchesPosTagRegex("POS:...:BAR"))
	require.True(t, tokenReadings.MatchesPosTagRegex("POS:[A-Z]+:BAR"))
	require.False(t, tokenReadings.MatchesPosTagRegex("POS:[AB]OO:BAR"))
	require.False(t, tokenReadings.MatchesPosTagRegex("POS:FOO:BARX"))
}

func TestAnalyzedTokenReadings_Iteration(t *testing.T) {
	tokenReadings := NewAnalyzedTokenReadingsList([]*AnalyzedToken{
		NewAnalyzedToken("word1", nil, nil),
		NewAnalyzedToken("word2", nil, nil),
	}, 0)
	i := 0
	for _, tokenReading := range tokenReadings.Readings() {
		if i == 0 {
			require.Equal(t, "word1", tokenReading.GetToken())
		} else if i == 1 {
			require.Equal(t, "word2", tokenReading.GetToken())
		} else {
			t.Fatal("unexpected iteration")
		}
		i++
	}
}

// Java toString appends chunk tags joined by '|' before the closing bracket.
func TestAnalyzedTokenReadings_ToString_ChunkTags(t *testing.T) {
	pos := "POS"
	lemma := "lemma"
	r := NewAnalyzedTokenReadings(NewAnalyzedToken("word", &pos, &lemma))
	r.SetChunkTags([]string{"B-NP", "NPP"})
	require.Equal(t, "word[lemma/POS*,B-NP|NPP]", r.String())
	r.Immunize(42)
	require.Equal(t, "word[lemma/POS*,B-NP|NPP]{!},", r.String())
}

// Java hasPosTagAndLemma: lemma.equals(reading.getLemma()) — null lemma never matches.
func TestAnalyzedTokenReadings_HasPosTagAndLemma_NullLemma(t *testing.T) {
	pos := "POS"
	r := NewAnalyzedTokenReadings(NewAnalyzedToken("word", &pos, nil))
	require.False(t, r.HasPosTagAndLemma("POS", ""))
	require.False(t, r.HasPosTagAndLemma("POS", "lemma"))
	lem := "lemma"
	r2 := NewAnalyzedTokenReadings(NewAnalyzedToken("word", &pos, &lem))
	require.True(t, r2.HasPosTagAndLemma("POS", "lemma"))
	require.False(t, r2.HasPosTagAndLemma("POS", "other"))
}

func TestAnalyzedTokenReadings_Equals(t *testing.T) {
	pos := "POS"
	lemma := "lemma"
	a := NewAnalyzedTokenReadings(NewAnalyzedToken("word", &pos, &lemma))
	b := NewAnalyzedTokenReadings(NewAnalyzedToken("word", &pos, &lemma))
	require.True(t, a.Equals(b))
	require.Equal(t, a.HashCode(), b.HashCode())
	a.SetChunkTags([]string{"B-NP"})
	require.False(t, a.Equals(b))
	b.SetChunkTags([]string{"B-NP"})
	require.True(t, a.Equals(b))
	a.IgnoreSpelling()
	require.False(t, a.Equals(b))
}
