package sr

// Twin of org.languagetool.synthesis.sr.JekavianSynthesizer (Java has no
// JekavianSynthesizerTest). Behavior matrix mirrors EkavianSynthesizerTest
// against the official jekavian serbian_synth.dict + shared serbian_synth_tags.txt.
// Expected forms are those returned by BaseSynthesizer + jekavian dict (same
// lemmas/tags as EkavianSynthesizerTest; jekavian dict yields these forms).

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func requireJekavianInstance(t *testing.T) *JekavianSynthesizer {
	t.Helper()
	s := JekavianInstance()
	if s == nil {
		t.Skip("serbian_synth.dict / official SR jekavian synth resources not available")
	}
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Same(t, s, JEKAVIAN_INSTANCE)
	return s
}

// Behavior twin of EkavianSynthesizerTest.testSynthesizeString run on JekavianSynthesizer.
func TestJekavianSynthesizer_SynthesizeString(t *testing.T) {
	synth := requireJekavianInstance(t)

	// Negative test - POS tag that does not exist
	forms, err := synth.Synthesize(dummyToken("катаклингер"), "катаклингер")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// Check all cases (same matrix as EkavianSynthesizerTest; jekavian dict forms)
	forms, err = synth.Synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:NO")
	require.NoError(t, err)
	require.Equal(t, "[оловка]", arraysToString(forms))

	forms, err = synth.Synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:GE")
	require.NoError(t, err)
	require.Equal(t, "[оловке]", arraysToString(forms))

	forms, err = synth.Synthesize(dummyToken("мајка"), "IM:ZA:ZE:0J:DA")
	require.NoError(t, err)
	require.Equal(t, "[мајци]", arraysToString(forms))

	forms, err = synth.Synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:AK")
	require.NoError(t, err)
	require.Equal(t, "[оловку]", arraysToString(forms))

	forms, err = synth.SynthesizeRE(dummyToken("пушка"), "IM:ZA:ZE:0J:VO", true)
	require.NoError(t, err)
	require.Equal(t, "[пушко]", arraysToString(forms))

	forms, err = synth.Synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:IN")
	require.NoError(t, err)
	require.Equal(t, "[оловком]", arraysToString(forms))

	forms, err = synth.Synthesize(dummyToken("оловка"), "IM:ZA:ZE:0J:LO")
	require.NoError(t, err)
	require.Equal(t, "[оловци]", arraysToString(forms))

	// regular expressions — sort before Arrays.toString (same as Ekavian test helper)
	forms, err = synth.SynthesizeRE(dummyToken("оловка"), "IM:ZA:ZE:0J:.*", true)
	require.NoError(t, err)
	require.Equal(t, "[оловка, оловка, оловке, оловко, оловком, оловку, оловци, оловци]", arraysToString(getSortedArray(forms)))

	forms, err = synth.SynthesizeRE(dummyToken("један"), "BR:.*:.*:.*:.*", true)
	require.NoError(t, err)
	require.Equal(t, "[један, један, један, једна, једна, једна, једна, једна, једне, једне, једне, једне, једни, једни, једни, једни, једним, једним, једним, једним, једним, једним, једним, једним, једним, једним, једним, једнима, једнима, једнима, једнима, једнима, једнима, једнима, једнима, једнима, једних, једних, једних, једно, једно, једно, једног, једног, једног, једнога, једнога, једнога, једном, једном, једном, једном, једном, једноме, једноме, једноме, једноме, једному, једному, једному, једному, једној, једној, једну]", arraysToString(getSortedArray(forms)))

	forms, err = synth.SynthesizeRE(dummyToken("орати"), "GL:.*:.*:.*:.*", true)
	require.NoError(t, err)
	require.Equal(t, "[орала, орала, орале, орали, орало, орао]", arraysToString(getSortedArray(forms)))

	forms, err = synth.SynthesizeRE(dummyToken("али"), "VE:.*", true)
	require.NoError(t, err)
	require.Equal(t, "[али]", arraysToString(getSortedArray(forms)))
}

// Constructor path constants (Java RESOURCE_FILENAME / TAGS_FILE_NAME).
func TestJekavianSynthesizer_ResourcePaths(t *testing.T) {
	j := NewJekavianSynthesizer(nil)
	require.Equal(t, JekavianSynthDict, j.ResourceFileName)
	require.Equal(t, TagsFileName, j.TagFileName)
	require.Equal(t, "sr", j.LangShortCode)
	require.Equal(t, "/sr/dictionary/serbian_synth_tags.txt", TagsFileName)
	require.Equal(t, "/sr/dictionary/jekavian/serbian_synth.dict", JekavianSynthDict)
	require.Equal(t, "/sr/dictionary/", DictionaryPath)
}

// Open path fail-closed + successful load from discovered jekavian dict.
func TestOpenJekavianSynthesizer_FromDict(t *testing.T) {
	require.Nil(t, OpenJekavianSynthesizerFromDictPath(""))
	require.Nil(t, OpenJekavianSynthesizerFromDictPath(t.TempDir()+"/nope.dict"))

	p := DiscoverJekavianSynthDict()
	if p == "" {
		t.Skip("jekavian serbian_synth.dict not discoverable")
	}
	s := OpenJekavianSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, JekavianSynthDict, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "sr", s.LangShortCode)
	// Shared tags file is loaded from dictionary/ (not jekavian/ only) — full no-dedupe list.
	require.Greater(t, len(s.PossibleTags), 1000)
}

// JekavianInstance singleton binds JEKAVIAN_INSTANCE to loaded official synthesizer.
func TestJekavianSynthesizer_Instance(t *testing.T) {
	s := JekavianInstance()
	if s == nil {
		t.Skip("serbian_synth.dict / official SR jekavian synth resources not available")
	}
	require.Same(t, s, JekavianInstance())
	require.Same(t, s, JEKAVIAN_INSTANCE)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, JekavianSynthDict, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
}
