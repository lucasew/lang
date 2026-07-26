package ar

// Twin of org.languagetool.synthesis.ar.ArabicSynthesizerTest using official
// ArabicSynthesizer.INSTANCE resources (arabic_synth.dict, arabic_tags.txt).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireArabicInstance(t *testing.T) *ArabicSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		t.Skip("arabic_synth.dict / official AR synth resources not available")
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

// Port of ArabicSynthesizerTest.testSynthesizeStringString
func TestArabicSynthesizer_SynthesizeStringString(t *testing.T) {
	synth := requireArabicInstance(t)

	// assertEquals(Arrays.toString(synth.synthesize(dummyToken("خيار"), "NJ-;F2--;---")), "[خيارتان, خياريتان]");
	forms, err := synth.Synthesize(dummyToken("خيار"), "NJ-;F2--;---")
	require.NoError(t, err)
	require.Equal(t, "[خيارتان, خياريتان]", arraysToString(forms))

	// assertEquals(Arrays.toString(synth.synthesize(dummyToken("بلاد"), "NJ-;F3A-;--H")),
	//   "[بلادت, بلادتي, بلاد, بلادي]");
	forms, err = synth.Synthesize(dummyToken("بلاد"), "NJ-;F3A-;--H")
	require.NoError(t, err)
	require.Equal(t, "[بلادت, بلادتي, بلاد, بلادي]", arraysToString(forms))

	// an example with specific postag with regex flag enabled
	// assertEquals(Arrays.toString(synth.synthesize(dummyToken("اِسْتَعْمَلَ"), "V61;M3Y-pa-;--H")),
	//   "[استعملتمو]");
	// Java calls synthesize(token, posTag) — two-arg, not RE (comment in source is misleading).
	forms, err = synth.Synthesize(dummyToken("اِسْتَعْمَلَ"), "V61;M3Y-pa-;--H")
	require.NoError(t, err)
	require.Equal(t, "[استعملتمو]", arraysToString(forms))

	// an example with specific postag without regex flag
	// assertEquals(Arrays.toString(synth.synthesize(dummyToken("اِسْتَكْمَلَ"), "V61;M3Y-pa-;---")),
	//   "[استكملتم]");
	forms, err = synth.Synthesize(dummyToken("اِسْتَكْمَلَ"), "V61;M3Y-pa-;---")
	require.NoError(t, err)
	require.Equal(t, "[استكملتم]", arraysToString(forms))

	// an example with specific postag without regex flag + code flag
	// assertEquals(Arrays.toString(synth.synthesize(dummyToken("اِسْتَمَعَ"), "V51;M3Y-pa-;---")),
	//   "[استمعتم]");
	forms, err = synth.Synthesize(dummyToken("اِسْتَمَعَ"), "V51;M3Y-pa-;---")
	require.NoError(t, err)
	require.Equal(t, "[استمعتم]", arraysToString(forms))
}

// Constructor path constants (Java RESOURCE_FILENAME / TAGS_FILE_NAME) without dict.
func TestArabicSynthesizer_ResourcePaths(t *testing.T) {
	s := NewArabicSynthesizer(nil)
	require.Equal(t, ArabicSynthDict, s.ResourceFileName)
	require.Equal(t, ArabicTagsFile, s.TagFileName)
	require.Equal(t, "ar", s.LangShortCode)
}

// Open path fail-closed + successful load from discovered dict.
func TestOpenArabicSynthesizer_FromDict(t *testing.T) {
	require.Nil(t, OpenArabicSynthesizerFromDir(""))
	require.Nil(t, OpenArabicSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenArabicSynthesizerFromDictPath(""))

	p := DiscoverArabicSynthDict()
	if p == "" {
		t.Skip("arabic_synth.dict not discoverable")
	}
	s := OpenArabicSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, ArabicSynthDict, s.ResourceFileName)
	require.Equal(t, ArabicTagsFile, s.TagFileName)
}

// Instance singleton binds INSTANCE to loaded official synthesizer.
func TestArabicSynthesizer_Instance(t *testing.T) {
	s := Instance()
	if s == nil {
		t.Skip("arabic_synth.dict / official AR synth resources not available")
	}
	require.Same(t, s, Instance())
	require.Same(t, s, INSTANCE)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
}
