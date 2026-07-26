package sv

// Twin of org.languagetool.synthesis.sv.SwedishSynthesizerTest using official
// SwedishSynthesizer.INSTANCE resources (swedish_synth.dict, swedish_synth.dict_tags.txt, sv.sor).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireSwedishInstance(t *testing.T) *SwedishSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		t.Skip("swedish_synth.dict / official SV synth resources not available")
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

// Port of SwedishSynthesizerTest.testSynthesizeStringString
func TestSwedishSynthesizer_SynthesizeStringString(t *testing.T) {
	synth := requireSwedishInstance(t)

	// assertEquals(synth.synthesize(dummyToken("blablabla"), "blablabla").length, 0);
	forms, err := synth.Synthesize(dummyToken("blablabla"), "blablabla")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// assertEquals("[nyckelns]", Arrays.toString(synth.synthesize(dummyToken("nyckel"), "NN:BF:SIN:GEN:UTR", true)));
	forms, err = synth.SynthesizeRE(dummyToken("nyckel"), "NN:BF:SIN:GEN:UTR", true)
	require.NoError(t, err)
	require.Equal(t, "[nyckelns]", arraysToString(forms))

	// assertEquals("[FN:s]", Arrays.toString(synth.synthesize(dummyToken("FN"), "PM:GEN:ACR", true)));
	forms, err = synth.SynthesizeRE(dummyToken("FN"), "PM:GEN:ACR", true)
	require.NoError(t, err)
	require.Equal(t, "[FN:s]", arraysToString(forms))

	// with regular expressions
	// assertEquals("[störste, största]", Arrays.toString(synth.synthesize(dummyToken("stor"), "JJ:S:BF:.*", true)));
	forms, err = synth.SynthesizeRE(dummyToken("stor"), "JJ:S:BF:.*", true)
	require.NoError(t, err)
	require.Equal(t, "[störste, största]", arraysToString(forms))

	// assertEquals("[VW:s, VW]", Arrays.toString(synth.synthesize(dummyToken("VW"), "PM:.*:ACR", true)));
	forms, err = synth.SynthesizeRE(dummyToken("VW"), "PM:.*:ACR", true)
	require.NoError(t, err)
	require.Equal(t, "[VW:s, VW]", arraysToString(forms))
}

// Constructor path constants (Java RESOURCE_FILENAME / TAGS_FILE_NAME / SOR_FILE_NAME) without dict.
func TestSwedishSynthesizer_ResourcePaths(t *testing.T) {
	s := NewSwedishSynthesizer(nil)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, SorFileName, s.SorFileName)
	require.Equal(t, "sv", s.LangShortCode)
}

// Open path fail-closed + successful load from discovered dict.
func TestOpenSwedishSynthesizer_FromDict(t *testing.T) {
	require.Nil(t, OpenSwedishSynthesizerFromDir(""))
	require.Nil(t, OpenSwedishSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenSwedishSynthesizerFromDictPath(""))

	p := DiscoverSwedishSynthDict()
	if p == "" {
		t.Skip("swedish_synth.dict not discoverable")
	}
	s := OpenSwedishSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, SorFileName, s.SorFileName)
}

// Instance singleton binds INSTANCE to loaded official synthesizer.
func TestSwedishSynthesizer_Instance(t *testing.T) {
	s := Instance()
	if s == nil {
		t.Skip("swedish_synth.dict / official SV synth resources not available")
	}
	require.Same(t, s, Instance())
	require.Same(t, s, INSTANCE)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
}
