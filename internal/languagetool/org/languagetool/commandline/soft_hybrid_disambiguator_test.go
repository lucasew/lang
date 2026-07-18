package commandline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/stretchr/testify/require"
)

func sentenceStartToken() *languagetool.AnalyzedTokenReadings {
	tag := languagetool.SentenceStartTagName
	return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil))
}

func TestRegisterSoftHybridDisambiguator_FR_Multiword(t *testing.T) {
	dir := t.TempDir()
	mw := filepath.Join(dir, "fr-multiwords.txt")
	require.NoError(t, os.WriteFile(mw, []byte("#separatorRegExp=[\\t;]\nhome page;N f s\n"), 0o644))
	lt := languagetool.NewJLanguageTool("fr")
	ok := RegisterSoftHybridDisambiguator(lt, "fr", SoftHybridPaths{Multiwords: mw})
	require.True(t, ok)
	require.NotNil(t, lt.Disambiguator)

	toks := []*languagetool.AnalyzedTokenReadings{
		sentenceStartToken(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("home", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("page", nil, nil)),
	}
	out := lt.Disambiguator.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.NotNil(t, out)
	found := false
	for _, tok := range out.GetTokens() {
		for _, r := range tok.GetReadings() {
			if r.GetPOSTag() == nil {
				continue
			}
			if strings.Contains(*r.GetPOSTag(), "N f s") {
				found = true
			}
		}
	}
	require.True(t, found, "FR multiword should apply after soft hybrid register")
}

func TestRegisterSoftHybridDisambiguator_PL_OrderXMLThenMulti(t *testing.T) {
	dir := t.TempDir()
	mw := filepath.Join(dir, "pl-multiwords.txt")
	require.NoError(t, os.WriteFile(mw, []byte("bez mała\tADV\n"), 0o644))
	lt := languagetool.NewJLanguageTool("pl")
	ok := RegisterSoftHybridDisambiguator(lt, "pl", SoftHybridPaths{Multiwords: mw})
	require.True(t, ok)
	require.NotNil(t, lt.Disambiguator)
}

func TestSoftHybridProfile_Orders(t *testing.T) {
	require.Equal(t, "global_mw_xml", softHybridProfile("fr").order)
	require.Equal(t, "xml_mw", softHybridProfile("pl").order)
	require.Equal(t, "xml_mw", softHybridProfile("sv").order)
	require.Equal(t, "mw_xml", softHybridProfile("ru").order)
	require.Equal(t, "de", softHybridProfile("de").order)
	require.True(t, softHybridProfile("fr").mwRemovePrev)
	require.Equal(t, disambiguation.TagForNotAddingTags, softHybridProfile("nl").mwDefaultTag)
	require.Equal(t, "NPCN000", softHybridProfile("es").gDefaultTag)
}

func TestDiscoverLanguageSoftHybridPaths(t *testing.T) {
	// Vendored soft FR resources should resolve from repo walk-up.
	xml := DiscoverLanguageSoftDisambiguationXML(nil, "fr")
	require.NotEmpty(t, xml, "fr soft disambig XML")
	mw := DiscoverLanguageMultiwords(nil, "fr")
	require.NotEmpty(t, mw, "fr multiwords")
}
