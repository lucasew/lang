package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDemo(t *testing.T) {
	d := NewDemo()
	require.Equal(t, DemoShortCode, d.GetShortCode())
	require.Equal(t, "Testlanguage", d.GetName())
	require.Equal(t, []string{"XX"}, d.GetCountries())
	require.NotNil(t, d.CreateDefaultTagger())
	require.NotNil(t, d.CreateDefaultWordTokenizer())
	require.NotNil(t, d.CreateDefaultChunker())
	require.NotNil(t, d.CreateDefaultDisambiguator())
	tags, err := d.CreateDefaultTagger().Tag([]string{"a"})
	require.NoError(t, err)
	require.Len(t, tags, 1)
	// chunker tags chunkbar
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("chunkbar", nil, nil)),
	}
	d.CreateDefaultChunker().AddChunkTags(toks)
	require.Equal(t, []string{"B-NP-singular"}, toks[0].GetChunkTags())
}

func TestMakeAdditionalLanguage(t *testing.T) {
	m, err := MakeAdditionalLanguage("/tmp/rules-de-German.xml")
	require.NoError(t, err)
	require.Equal(t, "de", m.Code)
	require.Equal(t, "German", m.Name)
	require.True(t, m.Additional)
	_, err = MakeAdditionalLanguage("bad.xml")
	require.Error(t, err)
}
