package uk

import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestPosTagHelper_Basic(t *testing.T) {
	require.True(t, IsNoun("noun:m:v_naz"))
	require.True(t, IsVerb("verb:imperf:inf"))
	require.Equal(t, "m", Gender("noun:m:v_naz"))
	require.Equal(t, "v_naz", Case("noun:m:v_naz"))
}

func TestPosTagHelper_GetGenderNumConj(t *testing.T) {
	// noun:inanim:m:v_naz
	require.Equal(t, "m", GetGender("noun:inanim:m:v_naz"))
	require.Equal(t, "s", GetNum("noun:inanim:m:v_naz"))
	require.Equal(t, "p", GetNum("noun:inanim:p:v_naz"))
	require.Equal(t, "v_naz", GetConj("noun:inanim:m:v_naz"))
	require.Equal(t, "m:v_naz", GetGenderConj("noun:inanim:m:v_naz"))
	require.Equal(t, "", GetGender("verb:imperf:inf"))
}

func TestPosTagHelper_HasPosTagFamily(t *testing.T) {
	pos := "noun:inanim:m:v_naz"
	tok := languagetool.NewAnalyzedToken("місто", &pos, strPtr("місто"))
	atr := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{tok}, 0)

	require.True(t, HasPosTagToken(tok, NounVNazPattern))
	require.True(t, HasPosTagReadings(atr, NounVNazPattern))
	require.True(t, HasPosTagPartToken(tok, "inanim"))
	require.True(t, HasPosTagPartReadings(atr, "v_naz"))
	require.True(t, HasPosTagStartToken(tok, "noun:inanim"))
	require.True(t, HasPosTagPartAllReadings(atr, "noun"))
	require.False(t, HasPosTagPartAllReadings(atr, "verb"))

	// SENT_END skipped in hasPosTagPartAll
	sentEnd := languagetool.SentenceEndTagName
	tok2 := languagetool.NewAnalyzedToken(".", &sentEnd, nil)
	atr2 := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{tok, tok2}, 0)
	require.True(t, HasPosTagPartAllReadings(atr2, "noun"))
}

func TestPosTagHelper_TaggedWordHelpers(t *testing.T) {
	words := []tagging.TaggedWord{
		tagging.NewTaggedWord("a", "adj:m:v_naz:compb"),
		tagging.NewTaggedWord("b", "noun:inanim:f:v_rod"),
	}
	require.True(t, HasPosTagPart2(words, "compb"))
	require.True(t, HasPosTag2(words, regexp.MustCompile(`^adj:.*$`)))
	require.True(t, HasPosTagStart2(words, "noun:"))

	adj := Filter2(words, regexp.MustCompile(`^adj:.*$`))
	require.Len(t, adj, 1)
	neg := Filter2Negative(words, regexp.MustCompile(`^adj:.*$`))
	require.Len(t, neg, 1)
	require.Equal(t, "b", neg[0].Lemma)
}

func TestPosTagHelper_AddAdjustGenerate(t *testing.T) {
	require.Equal(t, "adj:m:v_naz:bad", AddIfNotContains("adj:m:v_naz", ":bad"))
	require.Equal(t, "adj:m:v_naz:bad", AddIfNotContains("adj:m:v_naz:bad", ":bad"))
	require.Equal(t, "adj:m:v_naz:bad:alt", AddIfNotContainsMany("adj:m:v_naz", ":bad", ":alt"))

	words := []tagging.TaggedWord{tagging.NewTaggedWord("word", "adj:m:v_naz:compb")}
	adj := Adjust(words, "напів", "-x", ":bad")
	require.Equal(t, "напівword-x", adj[0].Lemma)
	require.Equal(t, "adj:m:v_naz:bad", adj[0].PosTag) // :compb cleaned

	nv := GenerateTokensForNv("Київ", "m", ":prop")
	// 6 cases (no v_kly) × 1 gender
	require.Len(t, nv, 6)
	require.Equal(t, "noun:inanim:m:v_naz:nv:prop", *nv[0].GetPOSTag())
	require.Equal(t, "Київ", *nv[0].GetLemma())
}

func TestPosTagHelper_PersonMapAndGenders(t *testing.T) {
	require.Equal(t, "одн.", PersonName("s"))
	require.Equal(t, "мн.", PersonName("p"))

	posM := "noun:inanim:m:v_naz"
	posF := "noun:inanim:f:v_naz"
	tokM := languagetool.NewAnalyzedToken("x", &posM, strPtr("x"))
	tokF := languagetool.NewAnalyzedToken("y", &posF, strPtr("y"))
	atr := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{tokM, tokF}, 0)
	g := GetGenders(atr, NounVNazPattern)
	require.Contains(t, g, "m")
	require.Contains(t, g, "f")
}

func TestPosTagHelper_IsUnknownAndPredict(t *testing.T) {
	// unknown: first reading hasNoTag + word-like surface
	empty := languagetool.NewAnalyzedToken("абракадабра", nil, nil)
	// HasNoTag typically true when pos is null
	atr := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{empty}, 0)
	if empty.HasNoTag() {
		require.True(t, IsUnknownWord(atr))
	}

	pos := "noninfl:predic"
	tok := languagetool.NewAnalyzedToken("треба", &pos, strPtr("треба"))
	require.True(t, IsPredictOrInsert(tok))
	pos2 := "noninfl:insert"
	tok2 := languagetool.NewAnalyzedToken("на жаль", &pos2, nil)
	require.True(t, IsPredictOrInsert(tok2))
	pos3 := "adv"
	tok3 := languagetool.NewAnalyzedToken("швидко", &pos3, nil)
	require.False(t, IsPredictOrInsert(tok3))
}

func TestPosTagHelper_HasMaleUA(t *testing.T) {
	pos := "noun:inanim:m:v_dav"
	tok := languagetool.NewAnalyzedToken("місту", &pos, strPtr("місто"))
	atr := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{tok}, 0)
	require.True(t, HasMaleUA(atr))

	posNV := "noun:inanim:m:v_dav:nv"
	tokNV := languagetool.NewAnalyzedToken("місту", &posNV, strPtr("місто"))
	atrNV := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{tokNV}, 0)
	require.False(t, HasMaleUA(atrNV))
}

func strPtr(s string) *string { return &s }
