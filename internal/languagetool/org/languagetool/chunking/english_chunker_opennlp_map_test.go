package chunking

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin of EnglishChunker.getAnalyzedTokenReadingsFor + getTokensWithTokenReadings:
// cumulative positions use String.length() (UTF-16); NBSP skipped via length==1+isSpaceChar.
func TestGetAnalyzedTokenReadingsFor_UTF16AndNBSP(t *testing.T) {
	// LT tokens: "café" + NBSP + "ok" — OpenNLP would emit ["café","ok"]
	cafe := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("café", nil, nil), 0)
	nbsp := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("\u00A0", nil, nil), 4)
	ok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("ok", nil, nil), 5)
	readings := []*languagetool.AnalyzedTokenReadings{cafe, nbsp, ok}

	// UTF-16: café = 4 units; after skip NBSP, "ok" is at 4..6
	got := getAnalyzedTokenReadingsFor(0, 4, readings)
	require.Same(t, cafe, got)
	got = getAnalyzedTokenReadingsFor(4, 6, readings)
	require.Same(t, ok, got, "NBSP must not advance cumulative UTF-16 pos")

	// Full map
	tagged := getTokensWithTokenReadings(readings, []string{"café", "ok"}, []string{"B-NP", "I-NP"})
	require.Len(t, tagged, 2)
	require.Same(t, cafe, tagged[0].Readings)
	require.Same(t, ok, tagged[1].Readings)
}

func TestJavaTrimIsEmpty_NotNBSP(t *testing.T) {
	// Java String.trim does not strip U+00A0
	require.False(t, javaTrimIsEmpty("\u00A0"))
	require.True(t, javaTrimIsEmpty("  \t"))
	require.True(t, javaTrimIsEmpty(""))
}
