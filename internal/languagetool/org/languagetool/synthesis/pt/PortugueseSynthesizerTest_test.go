package pt

// Twin of org.languagetool.synthesis.pt.PortugueseSynthesizerTest using official
// PortugueseSynthesizer.INSTANCE resources (portuguese_synth.dict, portuguese_tags.txt, pt.sor).

import (
	"sort"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requirePortugueseInstance(t *testing.T) *PortugueseSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		t.Skip("portuguese_synth.dict / official PT synth resources not available")
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

// Port of PortugueseSynthesizerTest.testSynthesizeString
func TestPortugueseSynthesizer_SynthesizeString(t *testing.T) {
	synth := requirePortugueseInstance(t)

	// assertEquals(synth.synthesize(dummyToken("blablabla"), "blablabla").length, 0);
	forms, err := synth.Synthesize(dummyToken("blablabla"), "blablabla")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// Java uses synthesize(token, posTag, true) — posTagRegExp path — then sorted Arrays.toString.
	// assertEquals("[bois]", Arrays.toString(getSortedArray(synth.synthesize(dummyToken("boi"), "NCMP000", true))));
	forms, err = synth.SynthesizeRE(dummyToken("boi"), "NCMP000", true)
	require.NoError(t, err)
	require.Equal(t, "[bois]", arraysToString(getSortedArray(forms)))

	// assertEquals("[tentar]", Arrays.toString(getSortedArray(synth.synthesize(dummyToken("tentar"), "VMN0000", true))));
	forms, err = synth.SynthesizeRE(dummyToken("tentar"), "VMN0000", true)
	require.NoError(t, err)
	require.Equal(t, "[tentar]", arraysToString(getSortedArray(forms)))

	// assertEquals("[tentou]", Arrays.toString(getSortedArray(synth.synthesize(dummyToken("tentar"), "VMIS3S0", true))));
	forms, err = synth.SynthesizeRE(dummyToken("tentar"), "VMIS3S0", true)
	require.NoError(t, err)
	require.Equal(t, "[tentou]", arraysToString(getSortedArray(forms)))

	// assertEquals("[resolver]", Arrays.toString(getSortedArray(synth.synthesize(dummyToken("resolver"), "VMN0000", true))));
	forms, err = synth.SynthesizeRE(dummyToken("resolver"), "VMN0000", true)
	require.NoError(t, err)
	require.Equal(t, "[resolver]", arraysToString(getSortedArray(forms)))

	// assertEquals("[resolveu]", Arrays.toString(getSortedArray(synth.synthesize(dummyToken("resolver"), "VMIS3S0", true))));
	forms, err = synth.SynthesizeRE(dummyToken("resolver"), "VMIS3S0", true)
	require.NoError(t, err)
	require.Equal(t, "[resolveu]", arraysToString(getSortedArray(forms)))

	// as of dict v1.0.1: issues with 'mesmíssimo' assigned to 'mesmo' lemma
	// assertEquals("[mesmas]", Arrays.toString(getSortedArray(synth.synthesize(dummyToken("mesmo"), "DD0FP0", true))));
	forms, err = synth.SynthesizeRE(dummyToken("mesmo"), "DD0FP0", true)
	require.NoError(t, err)
	require.Equal(t, "[mesmas]", arraysToString(getSortedArray(forms)))
}
