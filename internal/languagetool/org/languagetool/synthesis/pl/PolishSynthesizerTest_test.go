package pl

// Twin of org.languagetool.synthesis.pl.PolishSynthesizerTest using official
// PolishSynthesizer.INSTANCE resources (polish_synth.dict, polish_tags.txt).

import (
	"sort"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requirePolishInstance(t *testing.T) *PolishSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		t.Skip("polish_synth.dict / official PL synth resources not available")
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

// getSortedArray ports PolishSynthesizerTest.getSortedArray.
func getSortedArray(ar []string) []string {
	if ar == nil {
		return nil
	}
	newAr := append([]string(nil), ar...)
	sort.Strings(newAr)
	return newAr
}

// Port of PolishSynthesizerTest.testSynthesizeString — active asserts only
// (commented-out nieduży adj:sg:nom:m:pos:neg is not enabled).
func TestPolishSynthesizer_SynthesizeString(t *testing.T) {
	synth := requirePolishInstance(t)

	// assertEquals(synth.synthesize(dummyToken("blablabla"), "blablabla").length, 0);
	forms, err := synth.Synthesize(dummyToken("blablabla"), "blablabla")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// assertEquals("[Aarona]", Arrays.toString(synth.synthesize(dummyToken("Aaron"), "subst:sg:gen:m1")));
	forms, err = synth.Synthesize(dummyToken("Aaron"), "subst:sg:gen:m1")
	require.NoError(t, err)
	require.Equal(t, "[Aarona]", arraysToString(forms))

	// assertEquals("[Abchazem]", Arrays.toString(synth.synthesize(dummyToken("Abchaz"), "subst:sg:inst:m1")));
	forms, err = synth.Synthesize(dummyToken("Abchaz"), "subst:sg:inst:m1")
	require.NoError(t, err)
	require.Equal(t, "[Abchazem]", arraysToString(forms))

	// assertEquals("[miała]", Arrays.toString(synth.synthesize(dummyToken("mieć"), "verb:praet:sg:f:ter:imperf:refl.nonrefl")));
	forms, err = synth.Synthesize(dummyToken("mieć"), "verb:praet:sg:f:ter:imperf:refl.nonrefl")
	require.NoError(t, err)
	require.Equal(t, "[miała]", arraysToString(forms))

	// assertEquals("[brzydziej]", Arrays.toString(synth.synthesize(dummyToken("brzydko"), "adv:com")));
	forms, err = synth.Synthesize(dummyToken("brzydko"), "adv:com")
	require.NoError(t, err)
	require.Equal(t, "[brzydziej]", arraysToString(forms))

	// with regular expressions
	// assertEquals("[tonera]", Arrays.toString(getSortedArray(synth.synthesize(dummyToken("toner"), "subst:sg:gen:m.*", true))));
	forms, err = synth.SynthesizeRE(dummyToken("toner"), "subst:sg:gen:m.*", true)
	require.NoError(t, err)
	require.Equal(t, "[tonera]", arraysToString(getSortedArray(forms)))

	// assertEquals("[niedużego, nieduży]", Arrays.toString(getSortedArray(synth.synthesize(dummyToken("nieduży"), "adj:sg.*(m[0-9]?|m.n):pos", true))));
	forms, err = synth.SynthesizeRE(dummyToken("nieduży"), "adj:sg.*(m[0-9]?|m.n):pos", true)
	require.NoError(t, err)
	require.Equal(t, "[niedużego, nieduży]", arraysToString(getSortedArray(forms)))

	// assertEquals("[miał, miała, miałam, miałaś, miałem, miałeś, miało, miałom, miałoś]",
	//   Arrays.toString(getSortedArray(synth.synthesize(dummyToken("mieć"), ".*praet:sg.*", true))));
	forms, err = synth.SynthesizeRE(dummyToken("mieć"), ".*praet:sg.*", true)
	require.NoError(t, err)
	require.Equal(t,
		"[miał, miała, miałam, miałaś, miałem, miałeś, miało, miałom, miałoś]",
		arraysToString(getSortedArray(forms)))
}
