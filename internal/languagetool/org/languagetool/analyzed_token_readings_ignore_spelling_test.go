package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnalyzedTokenReadings_IgnoreSpelling(t *testing.T) {
	r := NewAnalyzedTokenReadings(NewAnalyzedToken("foo", nil, nil))
	require.False(t, r.IsIgnoredBySpeller())
	r.IgnoreSpelling()
	require.True(t, r.IsIgnoredBySpeller())

	// FromOld copies flag
	copy := NewAnalyzedTokenReadingsFromOld(r, []*AnalyzedToken{NewAnalyzedToken("foo", nil, nil)}, "rule")
	require.True(t, copy.IsIgnoredBySpeller())
}

// Java AnalyzedTokenReadings(old, readings, rule): setChunkTags(old.getChunkTags()).
func TestAnalyzedTokenReadings_FromOld_CopiesChunkTags(t *testing.T) {
	r := NewAnalyzedTokenReadings(NewAnalyzedToken("Autos", nil, nil))
	r.SetChunkTags([]string{"B-NP", "NPP"})
	copy := NewAnalyzedTokenReadingsFromOld(r, r.GetReadings(), "")
	require.Equal(t, []string{"B-NP", "NPP"}, copy.GetChunkTags())
}

// Java: if (oldAtr.hasTypographicApostrophe()) setTypographicApostrophe()
func TestAnalyzedTokenReadings_FromOld_CopiesTypographicApostrophe(t *testing.T) {
	r := NewAnalyzedTokenReadings(NewAnalyzedToken("l'eau", nil, nil))
	r.SetTypographicApostrophe(true)
	copy := NewAnalyzedTokenReadingsFromOld(r, r.GetReadings(), "")
	require.True(t, copy.HasTypographicApostrophe())
}

// Java historical annotations only when GlobalConfig.isVerbose()
func TestAnalyzedTokenReadings_FromOld_HistoricalAnnotationsVerbose(t *testing.T) {
	prev := IsVerbose()
	defer SetVerbose(prev)

	r := NewAnalyzedTokenReadings(NewAnalyzedToken("foo", nil, nil))
	SetVerbose(false)
	copyQuiet := NewAnalyzedTokenReadingsFromOld(r, r.GetReadings(), "RULE_X")
	require.Empty(t, copyQuiet.GetHistoricalAnnotations())

	SetVerbose(true)
	copyVerbose := NewAnalyzedTokenReadingsFromOld(r, r.GetReadings(), "RULE_X")
	require.Contains(t, copyVerbose.GetHistoricalAnnotations(), "RULE_X")
}

func TestAnalyzedTokenReadings_CleanTokenAndPosFix(t *testing.T) {
	r := NewAnalyzedTokenReadings(NewAnalyzedToken("soft\u00adhyphen", nil, nil))
	require.Equal(t, r.GetToken(), r.GetCleanToken())
	r.SetCleanToken("softhyphen")
	require.Equal(t, "softhyphen", r.GetCleanToken())
	r.SetPosFix(1)
	require.Equal(t, 1, r.GetPosFix())

	// GetCorrectedTextLength uses cleanToken + last posFix
	sent := NewAnalyzedSentence([]*AnalyzedTokenReadings{r})
	// "softhyphen" = 10 + posFix 1 = 11
	require.Equal(t, 11, sent.GetCorrectedTextLength())
}

func TestAnalyzedTokenReadings_IsPosTagUnknown(t *testing.T) {
	unk := NewAnalyzedTokenReadings(NewAnalyzedToken("xyz", nil, nil))
	require.True(t, unk.IsPosTagUnknown())
	p := "SUB:NOM:SIN:NEU"
	known := NewAnalyzedTokenReadings(NewAnalyzedToken("Haus", &p, strPtr("Haus")))
	require.False(t, known.IsPosTagUnknown())
}

func TestAnalyzedTokenReadings_ReadingWithTagRegexAndLemma(t *testing.T) {
	p1 := "VER:1:SIN:PRÄ:NON"
	p2 := "SUB:NOM:SIN:NEU"
	lem := "sein"
	r := NewAnalyzedTokenReadingsList([]*AnalyzedToken{
		NewAnalyzedToken("bin", &p1, &lem),
		NewAnalyzedToken("bin", &p2, strPtr("bin")),
	}, 0)
	require.NotNil(t, r.ReadingWithTagRegex(`VER:.*`))
	require.Equal(t, "VER:1:SIN:PRÄ:NON", *r.ReadingWithTagRegex(`VER:.*`).GetPOSTag())
	require.NotNil(t, r.ReadingWithLemma("sein"))
	require.True(t, r.HasLemma("sein"))
	require.True(t, r.HasAnyPartialPosTag("VER", "ADJ"))
	require.False(t, r.HasAnyPartialPosTag("EIG"))
	require.True(t, r.HasReading())
}

func strPtr(s string) *string { return &s }
