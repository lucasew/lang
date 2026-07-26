package sk

// Twin of org.languagetool.synthesis.sk.SlovakSynthesizerTest using official
// SlovakSynthesizer.INSTANCE resources (slovak_synth.dict, slovak_tags.txt, added.txt).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireSlovakInstance(t *testing.T) *SlovakSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		t.Skip("slovak_synth.dict / official SK synth resources not available")
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

// Port of SlovakSynthesizerTest.testSynthesizeStringString
func TestSlovakSynthesizer_SynthesizeStringString(t *testing.T) {
	synth := requireSlovakInstance(t)

	// assertEquals(synth.synthesize(dummyToken("blablabla"), "blablabla").length, 0);
	forms, err := synth.Synthesize(dummyToken("blablabla"), "blablabla")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// assertEquals("[časopisu]", Arrays.toString(synth.synthesize(dummyToken("časopis"), "SSis2")));
	forms, err = synth.Synthesize(dummyToken("časopis"), "SSis2")
	require.NoError(t, err)
	require.Equal(t, "[časopisu]", arraysToString(forms))

	// with regular expressions
	// assertEquals("[časopisy, časopisov, časopisom, časopisy, časopisy, časopisoch, časopismi, časopis, časopisu, časopisu, časopis, časopis, časopise, časopisom]", Arrays.toString(synth.synthesize(dummyToken("časopis"), "SS.*", true)));
	forms, err = synth.SynthesizeRE(dummyToken("časopis"), "SS.*", true)
	require.NoError(t, err)
	require.Equal(t, "[časopisy, časopisov, časopisom, časopisy, časopisy, časopisoch, časopismi, časopis, časopisu, časopisu, časopis, časopis, časopise, časopisom]", arraysToString(forms))
}

// Constructor path constants (Java RESOURCE_FILENAME / TAGS_FILE_NAME) without dict.
func TestSlovakSynthesizer_ResourcePaths(t *testing.T) {
	s := NewSlovakSynthesizer(nil)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "sk", s.LangShortCode)
}

// Open path fail-closed + successful load from discovered dict.
func TestOpenSlovakSynthesizer_FromDict(t *testing.T) {
	require.Nil(t, OpenSlovakSynthesizerFromDir(""))
	require.Nil(t, OpenSlovakSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenSlovakSynthesizerFromDictPath(""))

	p := DiscoverSlovakSynthDict()
	if p == "" {
		t.Skip("slovak_synth.dict not discoverable")
	}
	s := OpenSlovakSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
}

// Instance singleton binds INSTANCE to loaded official synthesizer.
func TestSlovakSynthesizer_Instance(t *testing.T) {
	s := Instance()
	if s == nil {
		t.Skip("slovak_synth.dict / official SK synth resources not available")
	}
	require.Same(t, s, Instance())
	require.Same(t, s, INSTANCE)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
}
