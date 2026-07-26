package it

// Twin of org.languagetool.synthesis.it.ItalianSynthesizerTest using official
// ItalianSynthesizer.INSTANCE resources (italian_synth.dict, italian_tags.txt).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireItalianInstance(t *testing.T) *ItalianSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		t.Skip("italian_synth.dict / official IT synth resources not available")
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

// Port of ItalianSynthesizerTest.testSynthesizeStringString
func TestItalianSynthesizer_SynthesizeStringString(t *testing.T) {
	synth := requireItalianInstance(t)

	// assertEquals(synth.synthesize(dummyToken("blablabla"), "blablabla").length, 0);
	forms, err := synth.Synthesize(dummyToken("blablabla"), "blablabla")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// assertEquals("[richiedi]", Arrays.toString(synth.synthesize(dummyToken("richiedere"), "VER:impr+pres+2+s")));
	forms, err = synth.Synthesize(dummyToken("richiedere"), "VER:impr+pres+2+s")
	require.NoError(t, err)
	require.Equal(t, "[richiedi]", arraysToString(forms))

	// assertEquals("[esultata]", Arrays.toString(synth.synthesize(dummyToken("esultare"), "VER:part+past+s+f")));
	forms, err = synth.Synthesize(dummyToken("esultare"), "VER:part+past+s+f")
	require.NoError(t, err)
	require.Equal(t, "[esultata]", arraysToString(forms))

	// assertEquals("[libri]", Arrays.toString(synth.synthesize(dummyToken("libro"), "NOUN-M:p")));
	forms, err = synth.Synthesize(dummyToken("libro"), "NOUN-M:p")
	require.NoError(t, err)
	require.Equal(t, "[libri]", arraysToString(forms))
}

// Constructor path constants (Java RESOURCE_FILENAME / TAGS_FILE_NAME) without dict.
func TestItalianSynthesizer_ResourcePaths(t *testing.T) {
	s := NewItalianSynthesizer(nil)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "it", s.LangShortCode)
}

// Open path fail-closed + successful load from discovered dict.
func TestOpenItalianSynthesizer_FromDict(t *testing.T) {
	require.Nil(t, OpenItalianSynthesizerFromDir(""))
	require.Nil(t, OpenItalianSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenItalianSynthesizerFromDictPath(""))

	p := DiscoverItalianSynthDict()
	if p == "" {
		t.Skip("italian_synth.dict not discoverable")
	}
	s := OpenItalianSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
}

// Instance singleton binds INSTANCE to loaded official synthesizer.
func TestItalianSynthesizer_Instance(t *testing.T) {
	s := Instance()
	if s == nil {
		t.Skip("italian_synth.dict / official IT synth resources not available")
	}
	require.Same(t, s, Instance())
	require.Same(t, s, INSTANCE)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
}
