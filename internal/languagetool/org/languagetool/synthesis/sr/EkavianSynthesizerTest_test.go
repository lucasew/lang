package sr

// Twin of org.languagetool.synthesis.sr.EkavianSynthesizerTest using official
// EkavianSynthesizer resources (serbian_synth.dict + serbian_synth_tags.txt).

import (
	"sort"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireEkavianInstance(t *testing.T) *EkavianSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		t.Skip("serbian_synth.dict / official SR ekavian synth resources not available")
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

// getSortedArray ports EkavianSynthesizerTest.getSortedArray (clone + Arrays.sort).
func getSortedArray(ar []string) []string {
	if ar == nil {
		return nil
	}
	newAr := append([]string(nil), ar...)
	sort.Strings(newAr)
	return newAr
}

// Port of EkavianSynthesizerTest.testSynthesizeString
func TestEkavianSynthesizer_SynthesizeString(t *testing.T) {
	synth := requireEkavianInstance(t)

	// Negative test - POS tag that does not exist
	// assertEquals(synth.synthesize(dummyToken("катаклингер"), "катаклингер").length, 0);
	forms, err := synth.Synthesize(dummyToken("катаклингер"), "катаклингер")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// Check all cases
	// assertEquals("[оловка]", Arrays.toString(synth.synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:NO")));
	forms, err = synth.Synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:NO")
	require.NoError(t, err)
	require.Equal(t, "[оловка]", arraysToString(forms))

	// assertEquals("[оловке]", Arrays.toString(synth.synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:GE")));
	forms, err = synth.Synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:GE")
	require.NoError(t, err)
	require.Equal(t, "[оловке]", arraysToString(forms))

	// assertEquals("[мајци]", Arrays.toString(synth.synthesize(dummyToken("мајка"), "IM:ZA:ZE:0J:DA")));
	forms, err = synth.Synthesize(dummyToken("мајка"), "IM:ZA:ZE:0J:DA")
	require.NoError(t, err)
	require.Equal(t, "[мајци]", arraysToString(forms))

	// assertEquals("[оловку]", Arrays.toString(synth.synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:AK")));
	forms, err = synth.Synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:AK")
	require.NoError(t, err)
	require.Equal(t, "[оловку]", arraysToString(forms))

	// assertEquals("[пушко]", Arrays.toString(synth.synthesize(dummyToken("пушка"), "IM:ZA:ZE:0J:VO", true)));
	forms, err = synth.SynthesizeRE(dummyToken("пушка"), "IM:ZA:ZE:0J:VO", true)
	require.NoError(t, err)
	require.Equal(t, "[пушко]", arraysToString(forms))

	// assertEquals("[оловком]", Arrays.toString(synth.synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:IN")));
	forms, err = synth.Synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:IN")
	require.NoError(t, err)
	require.Equal(t, "[оловком]", arraysToString(forms))

	// assertEquals("[оловци]", Arrays.toString(synth.synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:LO")));
	forms, err = synth.Synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:LO")
	require.NoError(t, err)
	require.Equal(t, "[оловци]", arraysToString(forms))

	// regular expressions — Java sorts before Arrays.toString via getSortedArray
	// assertEquals("[оловка, оловка, оловке, оловко, оловком, оловку, оловци, оловци]", ...);
	forms, err = synth.SynthesizeRE(dummyToken("оловка"), "IM:ZA:ZE:0J:.*", true)
	require.NoError(t, err)
	require.Equal(t, "[оловка, оловка, оловке, оловко, оловком, оловку, оловци, оловци]", arraysToString(getSortedArray(forms)))

	// assertEquals("[један, један, ... једну]", getSortedArray(synth.synthesize(dummyToken("један"), "BR:.*:.*:.*:.*", true)));
	forms, err = synth.SynthesizeRE(dummyToken("један"), "BR:.*:.*:.*:.*", true)
	require.NoError(t, err)
	require.Equal(t, "[један, један, један, једна, једна, једна, једна, једна, једне, једне, једне, једне, једни, једни, једни, једни, једним, једним, једним, једним, једним, једним, једним, једним, једним, једним, једним, једнима, једнима, једнима, једнима, једнима, једнима, једнима, једнима, једнима, једних, једних, једних, једно, једно, једно, једног, једног, једног, једнога, једнога, једнога, једном, једном, једном, једном, једном, једноме, једноме, једноме, једноме, једному, једному, једному, једному, једној, једној, једну]", arraysToString(getSortedArray(forms)))

	// assertEquals("[орала, орала, орале, орали, орало, орао]", ...);
	forms, err = synth.SynthesizeRE(dummyToken("орати"), "GL:.*:.*:.*:.*", true)
	require.NoError(t, err)
	require.Equal(t, "[орала, орала, орале, орали, орало, орао]", arraysToString(getSortedArray(forms)))

	// assertEquals("[али]", getSortedArray(synth.synthesize(dummyToken("али"), "VE:.*", true)));
	forms, err = synth.SynthesizeRE(dummyToken("али"), "VE:.*", true)
	require.NoError(t, err)
	require.Equal(t, "[али]", arraysToString(getSortedArray(forms)))
}

// Constructor path constants (Java RESOURCE / TAGS) without dict.
func TestEkavianSynthesizer_ResourcePaths(t *testing.T) {
	s := NewEkavianSynthesizer(nil)
	require.Equal(t, EkavianSynthDict, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "sr", s.LangShortCode)
	require.Equal(t, "/sr/dictionary/serbian_synth_tags.txt", TagsFileName)
	require.Equal(t, "/sr/dictionary/ekavian/serbian_synth.dict", EkavianSynthDict)

	j := NewJekavianSynthesizer(nil)
	require.Equal(t, JekavianSynthDict, j.ResourceFileName)
	require.Equal(t, TagsFileName, j.TagFileName)
	require.Equal(t, "sr", j.LangShortCode)
}

// Open path fail-closed + successful load from discovered dict.
func TestOpenEkavianSynthesizer_FromDict(t *testing.T) {
	require.Nil(t, OpenEkavianSynthesizerFromDictPath(""))
	require.Nil(t, OpenEkavianSynthesizerFromDictPath(t.TempDir()+"/nope.dict"))

	p := DiscoverEkavianSynthDict()
	if p == "" {
		t.Skip("serbian_synth.dict not discoverable")
	}
	s := OpenEkavianSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, EkavianSynthDict, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "sr", s.LangShortCode)
}

// Instance singleton binds INSTANCE to loaded official synthesizer.
func TestEkavianSynthesizer_Instance(t *testing.T) {
	s := Instance()
	if s == nil {
		t.Skip("serbian_synth.dict / official SR ekavian synth resources not available")
	}
	require.Same(t, s, Instance())
	require.Same(t, s, INSTANCE)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
}
