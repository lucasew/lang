package de

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpellingAdjExpansion_A(t *testing.T) {
	ex, err := LoadSpellingAdjExpansion(strings.NewReader("abknickbar/A\n"))
	require.NoError(t, err)
	// base
	tags := ex.Tag("abknickbar")
	require.NotEmpty(t, tags)
	require.Equal(t, "abknickbar", tags[0].Lemma)
	require.Equal(t, "ADJ:PRD:GRU", tags[0].PosTag)
	// -e form
	tagsE := ex.Tag("abknickbare")
	require.NotEmpty(t, tagsE)
	require.True(t, strings.HasPrefix(tagsE[0].PosTag, "ADJ:"))
	// ste/A skipped
	ex2, err := LoadSpellingAdjExpansion(strings.NewReader("fünftjünste/A\n"))
	require.NoError(t, err)
	require.Equal(t, 0, ex2.Size())
}

func TestSpellingAdjExpansion_P_ToPA2(t *testing.T) {
	ex, err := LoadSpellingAdjExpansion(strings.NewReader("abgemindert/P\n"))
	require.NoError(t, err)
	tags := ex.Tag("abgemindert")
	require.NotEmpty(t, tags)
	require.True(t, strings.HasPrefix(tags[0].PosTag, "PA2:"), tags[0].PosTag)
	require.True(t, strings.HasSuffix(tags[0].PosTag, ":VER"), tags[0].PosTag)
	tagsE := ex.Tag("abgeminderte")
	require.NotEmpty(t, tagsE)
	require.True(t, strings.HasPrefix(tagsE[0].PosTag, "PA2:"))
	require.True(t, strings.HasSuffix(tagsE[0].PosTag, ":VER"))
}

func TestSpellingAdjExpansion_FromOfficialFile(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	path := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/spelling.txt")
	ex, err := LoadSpellingAdjExpansionFromFile(path)
	if err != nil {
		t.Skipf("spelling.txt not found: %v", err)
	}
	require.Greater(t, ex.Size(), 1000)
	// known /A line abknickbar
	require.NotEmpty(t, ex.Tag("abknickbar"))
	// known /P line
	require.NotEmpty(t, ex.Tag("abgemindert"))
	require.True(t, strings.HasPrefix(ex.Tag("abgemindert")[0].PosTag, "PA2:"))
}

func TestGermanTagger_AdjExpansion(t *testing.T) {
	ex, err := LoadSpellingAdjExpansion(strings.NewReader("abknickbar/A\n"))
	require.NoError(t, err)
	tagger := NewGermanTagger(nil)
	tagger.SetSpellingAdjExpansion(ex)
	rd := tagger.Lookup("abknickbare")
	require.NotNil(t, rd)
	require.NotNil(t, rd.GetReadings()[0].GetPOSTag())
	require.True(t, strings.HasPrefix(*rd.GetReadings()[0].GetPOSTag(), "ADJ:"))
}

// Twin: fillAdjInfos put replaces — second line for same fullform wins (no append).
func TestSpellingAdjExpansion_PutReplaces(t *testing.T) {
	ex, err := LoadSpellingAdjExpansion(strings.NewReader("foo/A\nfoo/P\n"))
	require.NoError(t, err)
	tags := ex.Tag("foo")
	require.NotEmpty(t, tags)
	// last wins: /P → PA2
	require.True(t, strings.HasPrefix(tags[0].PosTag, "PA2:"), tags[0].PosTag)
	// exact count of tags for base form = len(TagsForAdj) after ToPA2, not double
	require.Equal(t, len(ToPA2(TagsForAdj)), len(tags))
}

// Twin: er/A and ste/A skipped
func TestSpellingAdjExpansion_SkipComparative(t *testing.T) {
	ex, err := LoadSpellingAdjExpansion(strings.NewReader("stärker/A\nschönste/A\n"))
	require.NoError(t, err)
	require.Equal(t, 0, ex.Size())
}

// Twin: base = replaceFirst("/.*","")
func TestSpellingAdjExpansion_StripFromFirstSlash(t *testing.T) {
	ex, err := LoadSpellingAdjExpansion(strings.NewReader("bar/A\n"))
	require.NoError(t, err)
	require.NotEmpty(t, ex.Tag("bar"))
	require.Equal(t, "bar", ex.Tag("bar")[0].Lemma)
}
