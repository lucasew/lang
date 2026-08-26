package ru

// Twin of org.languagetool.synthesis.ru.RussianSynthesizerTest using official
// RussianSynthesizer.INSTANCE resources (russian_synth.dict, tags_russian.txt).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireRussianInstance(t *testing.T) *RussianSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		t.Skip("russian_synth.dict / official RU synth resources not available")
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

// Port of RussianSynthesizerTest.testSynthesizeString
func TestRussianSynthesizer_SynthesizeString(t *testing.T) {
	synth := requireRussianInstance(t)

	// assertEquals(synth.synthesize(dummyToken("blablabla"), "blablabla").length, 0);
	forms, err := synth.Synthesize(dummyToken("blablabla"), "blablabla")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// assertEquals("[семья]", Arrays.toString(synth.synthesize(dummyToken("семья"), "NN:Inanim:Fem:Sin:Nom")));
	forms, err = synth.Synthesize(dummyToken("семья"), "NN:Inanim:Fem:Sin:Nom")
	require.NoError(t, err)
	require.Equal(t, "[семья]", arraysToString(forms))

	// assertEquals("[семьи]", Arrays.toString(synth.synthesize(dummyToken("семья"), "NN:Inanim:Fem:Sin:R")));
	forms, err = synth.Synthesize(dummyToken("семья"), "NN:Inanim:Fem:Sin:R")
	require.NoError(t, err)
	require.Equal(t, "[семьи]", arraysToString(forms))
}

// Constructor path constants (Java RESOURCE_FILENAME / TAGS_FILE_NAME) without dict.
func TestRussianSynthesizer_ResourcePaths(t *testing.T) {
	s := NewRussianSynthesizer(nil)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "ru", s.LangShortCode)
}

// Open path fail-closed + successful load from discovered dict.
func TestOpenRussianSynthesizer_FromDict(t *testing.T) {
	require.Nil(t, OpenRussianSynthesizerFromDir(""))
	require.Nil(t, OpenRussianSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenRussianSynthesizerFromDictPath(""))

	p := DiscoverRussianSynthDict()
	if p == "" {
		t.Skip("russian_synth.dict not discoverable")
	}
	s := OpenRussianSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
}
