package ga

// Outcome twins for org.languagetool.synthesis.ga.IrishSynthesizer using official
// IrishSynthesizer.INSTANCE resources (irish_synth.dict, irish_tags.txt, added.txt).
// Java has no IrishSynthesizerTest; cases assert BaseSynthesizer outcomes on real dict
// (lemma|tag pairs probed from official irish_synth.dict — no invent forms).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireIrishInstance(t *testing.T) *IrishSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		t.Skip("irish_synth.dict / official GA synth resources not available")
	}
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Same(t, s, INSTANCE)
	return s
}

func dummyToken(tokenStr string) *languagetool.AnalyzedToken {
	// Java: new AnalyzedToken(tokenStr, tokenStr, tokenStr)
	tag := tokenStr
	lem := tokenStr
	return languagetool.NewAnalyzedToken(tokenStr, &tag, &lem)
}

// arraysToString ports Java Arrays.toString(String[]) for twin assertions.
func arraysToString(forms []string) string {
	if forms == nil {
		return "[]"
	}
	return "[" + strings.Join(forms, ", ") + "]"
}

// BaseSynthesizer outcomes on real irish_synth.dict (Java-visible forms).
func TestIrishSynthesizer_SynthesizeStringString(t *testing.T) {
	synth := requireIrishInstance(t)

	// unknown lemma → empty (same pattern as RO/IT/UK Java tests)
	forms, err := synth.Synthesize(dummyToken("blablabla"), "blablabla")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// bí — present indicative (analytic)
	// official dict: bí|Verb:VI:PresInd → tá
	forms, err = synth.Synthesize(dummyToken("bí"), "Verb:VI:PresInd")
	require.NoError(t, err)
	require.Equal(t, "[tá]", arraysToString(forms))

	// bí — future
	// official dict: bí|Verb:VI:FutInd → beidh
	forms, err = synth.Synthesize(dummyToken("bí"), "Verb:VI:FutInd")
	require.NoError(t, err)
	require.Equal(t, "[beidh]", arraysToString(forms))

	// bí — 1st person singular present
	// official dict: bí|Verb:VI:PresInd:1P:Sg → táim
	forms, err = synth.Synthesize(dummyToken("bí"), "Verb:VI:PresInd:1P:Sg")
	require.NoError(t, err)
	require.Equal(t, "[táim]", arraysToString(forms))

	// fear — common singular / genitive / plural
	forms, err = synth.Synthesize(dummyToken("fear"), "Noun:Masc:Com:Sg")
	require.NoError(t, err)
	require.Equal(t, "[fear]", arraysToString(forms))

	forms, err = synth.Synthesize(dummyToken("fear"), "Noun:Masc:Gen:Sg")
	require.NoError(t, err)
	require.Equal(t, "[fir]", arraysToString(forms))

	forms, err = synth.Synthesize(dummyToken("fear"), "Noun:Masc:Com:Pl")
	require.NoError(t, err)
	require.Equal(t, "[fir]", arraysToString(forms))

	// fear — lenited
	forms, err = synth.Synthesize(dummyToken("fear"), "Noun:Masc:Com:Sg:Len")
	require.NoError(t, err)
	require.Equal(t, "[fhear]", arraysToString(forms))

	// bean — common singular / plural
	forms, err = synth.Synthesize(dummyToken("bean"), "Noun:Fem:Com:Sg")
	require.NoError(t, err)
	require.Equal(t, "[bean]", arraysToString(forms))

	forms, err = synth.Synthesize(dummyToken("bean"), "Noun:Fem:Com:Pl")
	require.NoError(t, err)
	require.Equal(t, "[mná]", arraysToString(forms))

	// maith — base / lenited
	forms, err = synth.Synthesize(dummyToken("maith"), "Adj:Base")
	require.NoError(t, err)
	require.Equal(t, "[maith]", arraysToString(forms))

	forms, err = synth.Synthesize(dummyToken("maith"), "Adj:Base:Len")
	require.NoError(t, err)
	require.Equal(t, "[mhaith]", arraysToString(forms))

	// RegExp — BaseSynthesizer.SynthesizeRE over possible tags (fear Com Sg variants)
	forms, err = synth.SynthesizeRE(dummyToken("fear"), "Noun:Masc:Com:Sg", true)
	require.NoError(t, err)
	require.Contains(t, forms, "fear")

	forms, err = synth.SynthesizeRE(dummyToken("fear"), "Noun:Masc:Com:Sg:Len", true)
	require.NoError(t, err)
	require.Equal(t, "[fhear]", arraysToString(forms))

	// user data (/ga/added.txt) — fata from official added.txt
	// fullform baseform postags: fata\tfata\tNoun:Masc:Com:Sg
	forms, err = synth.Synthesize(dummyToken("fata"), "Noun:Masc:Com:Sg")
	require.NoError(t, err)
	require.Equal(t, "[fata]", arraysToString(forms))
}

// Constructor path constants (Java RESOURCE_FILENAME / TAGS_FILE_NAME) without dict.
func TestIrishSynthesizer_ResourcePaths(t *testing.T) {
	s := NewIrishSynthesizer(nil)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "ga", s.LangShortCode)
}

// Open path fail-closed + successful load from discovered dict.
func TestOpenIrishSynthesizer_FromDict(t *testing.T) {
	require.Nil(t, OpenIrishSynthesizerFromDir(""))
	require.Nil(t, OpenIrishSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenIrishSynthesizerFromDictPath(""))

	p := DiscoverIrishSynthDict()
	if p == "" {
		t.Skip("irish_synth.dict not discoverable")
	}
	s := OpenIrishSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
}

// Instance singleton binds INSTANCE to loaded official synthesizer.
func TestIrishSynthesizer_Instance(t *testing.T) {
	s := Instance()
	if s == nil {
		t.Skip("irish_synth.dict / official GA synth resources not available")
	}
	require.Same(t, s, Instance())
	require.Same(t, s, INSTANCE)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
}
