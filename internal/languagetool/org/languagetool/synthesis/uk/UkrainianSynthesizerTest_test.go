package uk

// Twin of org.languagetool.synthesis.uk.UkrainianSynthesizerTest using official
// UkrainianSynthesizer.INSTANCE resources (ukrainian_synth.dict, ukrainian_tags.txt).

import (
	"sort"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireUkrainianInstance(t *testing.T) *UkrainianSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		t.Skip("ukrainian_synth.dict / official UK synth resources not available")
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

func getSortedArray(ar []string) []string {
	// Java: ar.clone(); Arrays.sort(newAr);
	if ar == nil {
		return nil
	}
	out := append([]string(nil), ar...)
	sort.Strings(out)
	return out
}

// Port of UkrainianSynthesizerTest.testSynthesizeString
func TestUkrainianSynthesizer_SynthesizeString(t *testing.T) {
	synth := requireUkrainianInstance(t)

	// assertEquals(synth.synthesize(dummyToken("щосьтамтаке"), "щосьтамтаке").length, 0);
	forms, err := synth.Synthesize(dummyToken("щосьтамтаке"), "щосьтамтаке")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// assertEquals("[міста]", Arrays.toString(synth.synthesize(dummyToken("місто"), "noun:inanim:n:v_rod")));
	forms, err = synth.Synthesize(dummyToken("місто"), "noun:inanim:n:v_rod")
	require.NoError(t, err)
	require.Equal(t, "[міста]", arraysToString(forms))

	// with regular expressions (sorted for assertEquals)
	forms, err = synth.SynthesizeRE(dummyToken("червоний"), "adj:f:v_rod.*", true)
	require.NoError(t, err)
	require.Equal(t, "[червоної]", arraysToString(getSortedArray(forms)))

	forms, err = synth.SynthesizeRE(dummyToken("червоний"), "adj:f:v_rod:compb", true)
	require.NoError(t, err)
	require.Equal(t, "[червоної]", arraysToString(getSortedArray(forms)))

	forms, err = synth.SynthesizeRE(dummyToken("червоний"), "adj:.:v_oru:compb", true)
	require.NoError(t, err)
	require.Equal(t, "[червоним, червоним, червоними, червоною]", arraysToString(getSortedArray(forms)))
}

// Constructor path constants (Java RESOURCE_FILENAME / TAGS_FILE_NAME) without dict.
func TestUkrainianSynthesizer_ResourcePaths(t *testing.T) {
	s := NewUkrainianSynthesizer(nil)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "uk", s.LangShortCode)
}

// Open path fail-closed + successful load from discovered dict.
func TestOpenUkrainianSynthesizer_FromDict(t *testing.T) {
	require.Nil(t, OpenUkrainianSynthesizerFromDir(""))
	require.Nil(t, OpenUkrainianSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenUkrainianSynthesizerFromDictPath(""))

	p := DiscoverUkrainianSynthDict()
	if p == "" {
		t.Skip("ukrainian_synth.dict not discoverable")
	}
	s := OpenUkrainianSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
}
