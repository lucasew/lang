package ro

// Twin of org.languagetool.synthesis.ro.RomanianSynthesizerTest using official
// RomanianSynthesizer.INSTANCE resources (romanian_synth.dict, romanian_tags.txt, added.txt).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireRomanianInstance(t *testing.T) *RomanianSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		t.Skip("romanian_synth.dict / official RO synth resources not available")
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

// Port of RomanianSynthesizerTest.testSynthesizeStringString
func TestRomanianSynthesizer_SynthesizeStringString(t *testing.T) {
	synth := requireRomanianInstance(t)

	// assertEquals(synth.synthesize(dummyToken("blablabla"), "blablabla").length, 0);
	forms, err := synth.Synthesize(dummyToken("blablabla"), "blablabla")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// a alege — forma de infinitiv
	// assertEquals("[alege]", Arrays.toString(synth.synthesize(dummyToken("alege"), "V000000f00")));
	forms, err = synth.Synthesize(dummyToken("alege"), "V000000f00")
	require.NoError(t, err)
	require.Equal(t, "[alege]", arraysToString(forms))

	// conjunctiv, pers a doua plural
	// assertEquals("[alegeți]", Arrays.toString(synth.synthesize(dummyToken("alege"), "V0p2000cz0")));
	forms, err = synth.Synthesize(dummyToken("alege"), "V0p2000cz0")
	require.NoError(t, err)
	require.Equal(t, "[alegeți]", arraysToString(forms))

	// a fi
	// assertEquals("[fi]", Arrays.toString(synth.synthesize(dummyToken("fi"), "V000000f0f")));
	forms, err = synth.Synthesize(dummyToken("fi"), "V000000f0f")
	require.NoError(t, err)
	require.Equal(t, "[fi]", arraysToString(forms))

	// indicativ prezent, pers a doua plural
	// assertEquals("[sunteți]", Arrays.toString(synth.synthesize(dummyToken("fi"), "V0p2000izf")));
	forms, err = synth.Synthesize(dummyToken("fi"), "V0p2000izf")
	require.NoError(t, err)
	require.Equal(t, "[sunteți]", arraysToString(forms))

	// indicativ prezent, pers a treia plural
	// assertEquals("[sunt]", Arrays.toString(synth.synthesize(dummyToken("fi"), "V0p3000izf")));
	forms, err = synth.Synthesize(dummyToken("fi"), "V0p3000izf")
	require.NoError(t, err)
	require.Equal(t, "[sunt]", arraysToString(forms))

	// indicativ prezent, pers întâi plural
	// assertEquals("[sunt]", Arrays.toString(synth.synthesize(dummyToken("fi"), "V0s1000izf")));
	forms, err = synth.Synthesize(dummyToken("fi"), "V0s1000izf")
	require.NoError(t, err)
	require.Equal(t, "[sunt]", arraysToString(forms))

	// RegExp — indicativ prezent, pers a doua plural SAU pers a treia plural
	// assertEquals("[sunteți, sunt]", Arrays.toString(synth.synthesize(dummyToken("fi"), "V0p2000izf|V0p3000izf", true)));
	forms, err = synth.SynthesizeRE(dummyToken("fi"), "V0p2000izf|V0p3000izf", true)
	require.NoError(t, err)
	require.Equal(t, "[sunteți, sunt]", arraysToString(forms))

	// diverse — indicativ, mai mult ca perfect, persoana întâi, plural
	// assertEquals("[merseserăm]", Arrays.toString(synth.synthesize(dummyToken("merge"), "V0p1000im0")));
	forms, err = synth.Synthesize(dummyToken("merge"), "V0p1000im0")
	require.NoError(t, err)
	require.Equal(t, "[merseserăm]", arraysToString(forms))

	// indicativ, mai mult ca perfect, persoana întâi, singular
	// assertEquals("[mersesem]", Arrays.toString(synth.synthesize(dummyToken("merge"), "V0s1000im0")));
	forms, err = synth.Synthesize(dummyToken("merge"), "V0s1000im0")
	require.NoError(t, err)
	require.Equal(t, "[mersesem]", arraysToString(forms))

	// assertEquals("[legătura]", Arrays.toString(synth.synthesize(dummyToken("legătură"), "Sfs3aac000")));
	forms, err = synth.Synthesize(dummyToken("legătură"), "Sfs3aac000")
	require.NoError(t, err)
	require.Equal(t, "[legătura]", arraysToString(forms))

	// assertEquals("[legătură]", Arrays.toString(synth.synthesize(dummyToken("legătură"), "Sfs3anc000")));
	forms, err = synth.Synthesize(dummyToken("legătură"), "Sfs3anc000")
	require.NoError(t, err)
	require.Equal(t, "[legătură]", arraysToString(forms))

	// user data (/ro/added.txt)
	// assertEquals("[configurați]", Arrays.toString(synth.synthesize(dummyToken("configura"), "V0p2000cz0")));
	forms, err = synth.Synthesize(dummyToken("configura"), "V0p2000cz0")
	require.NoError(t, err)
	require.Equal(t, "[configurați]", arraysToString(forms))

	// assertEquals("[configurați, configurezi]", Arrays.toString(synth.synthesize(dummyToken("configura"), "V0.2000cz0", true)));
	forms, err = synth.SynthesizeRE(dummyToken("configura"), "V0.2000cz0", true)
	require.NoError(t, err)
	require.Equal(t, "[configurați, configurezi]", arraysToString(forms))
}

// Constructor path constants (Java RESOURCE_FILENAME / TAGS_FILE_NAME) without dict.
func TestRomanianSynthesizer_ResourcePaths(t *testing.T) {
	s := NewRomanianSynthesizer(nil)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "ro", s.LangShortCode)
}

// Open path fail-closed + successful load from discovered dict.
func TestOpenRomanianSynthesizer_FromDict(t *testing.T) {
	require.Nil(t, OpenRomanianSynthesizerFromDir(""))
	require.Nil(t, OpenRomanianSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenRomanianSynthesizerFromDictPath(""))

	p := DiscoverRomanianSynthDict()
	if p == "" {
		t.Skip("romanian_synth.dict not discoverable")
	}
	s := OpenRomanianSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
}

// Instance singleton binds INSTANCE to loaded official synthesizer.
func TestRomanianSynthesizer_Instance(t *testing.T) {
	s := Instance()
	if s == nil {
		t.Skip("romanian_synth.dict / official RO synth resources not available")
	}
	require.Same(t, s, Instance())
	require.Same(t, s, INSTANCE)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
}
