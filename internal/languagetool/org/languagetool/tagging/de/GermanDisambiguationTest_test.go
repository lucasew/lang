package de

// Twin of GermanDisambiguationTest — chunker + MultitokenIgnore from official lists
import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disde "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/de"
	"github.com/stretchr/testify/require"
)

func atrDE(token, pos string) *languagetool.AnalyzedTokenReadings {
	p := pos
	return languagetool.NewAnalyzedTokenReadingsList(
		[]*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(token, &p, nil)}, 0)
}

func sentStartDE() *languagetool.AnalyzedTokenReadings {
	// SENT_START so MultiWordChunker space branch (j > 1) matches Java indexing.
	tag := languagetool.SentenceStartTagName
	return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil))
}

// openOfficialMultitokenIgnore loads Java /de/multitoken-ignore.txt (no invent patterns).
func openOfficialMultitokenIgnore(t *testing.T) *disambiguation.MultiWordChunker {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	path := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/multitoken-ignore.txt")
	if st, err := os.Stat(path); err != nil || !st.Mode().IsRegular() {
		t.Skipf("official multitoken-ignore.txt missing: %s", path)
	}
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	c, err := disambiguation.NewMultiWordChunkerFromReader(f, disambiguation.MultiWordChunkerSettings{
		AllowFirstCapitalized: true,
		AllowAllUppercase:     true,
		AllowTitlecase:        false,
		DefaultTag:            disambiguation.TagForNotAddingTags,
	})
	require.NoError(t, err)
	require.NotNil(t, c)
	c.AddIgnoreSpelling = true
	return c
}

// Port of GermanDisambiguationTest.testChunker
func TestGermanDisambiguation_Chunker(t *testing.T) {
	toks2 := []*languagetool.AnalyzedTokenReadings{
		atrDE("ein", "ART:IND:AKK:SIN:NEU"),
		atrDE("Konzept", "SUB:AKK:SIN:NEU"),
	}
	ch := chunking.NewGermanChunker()
	ch.AddChunkTags(toks2)
	// REGEXES2 may add NPS on top of B-NP/I-NP (Java additive tags).
	require.Contains(t, toks2[0].GetChunkTags(), "B-NP")
	require.Contains(t, toks2[1].GetChunkTags(), "I-NP")

	// Java MultitokenIgnore: multitoken-ignore.txt "3-adische System/E"
	// (not invent digit-hyphen regex on bare "3-adische").
	mw := openOfficialMultitokenIgnore(t)
	d := disde.NewGermanRuleDisambiguator()
	d.MultitokenIgnore = mw

	toks3 := []*languagetool.AnalyzedTokenReadings{
		sentStartDE(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("3-adische", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("System", nil, nil)),
	}
	out := d.Disambiguate(languagetool.NewAnalyzedSentence(toks3))
	require.True(t, out.GetTokens()[1].IsIgnoredBySpeller(), "3-adische from multitoken-ignore")
	require.True(t, out.GetTokens()[3].IsIgnoredBySpeller(), "System multiword span ignore")

	// Kelassurier Mauer is in multitoken-ignore.txt — both tokens ignored.
	toks4 := []*languagetool.AnalyzedTokenReadings{
		sentStartDE(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Kelassurier", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Mauer", nil, nil)),
	}
	out4 := d.Disambiguate(languagetool.NewAnalyzedSentence(toks4))
	require.True(t, out4.GetTokens()[1].IsIgnoredBySpeller(), "Kelassurier from multitoken-ignore")
	require.True(t, out4.GetTokens()[3].IsIgnoredBySpeller(), "Mauer multiword span ignore")

	// Bare Kelassurier alone is not a multitoken match — no invent single-token ignore.
	toks5 := []*languagetool.AnalyzedTokenReadings{
		sentStartDE(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Kelassurier", nil, nil)),
	}
	out5 := d.Disambiguate(languagetool.NewAnalyzedSentence(toks5))
	require.False(t, out5.GetTokens()[1].IsIgnoredBySpeller(), "single Kelassurier needs multiword partner")
}
