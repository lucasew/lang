package fr

// Twin of org.languagetool.synthesis.fr.FrenchSynthesizerTest using official
// FrenchSynthesizer.INSTANCE resources (french_synth.dict, french_tags.txt, added/removed, fr.sor).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireFrenchInstance(t *testing.T) *FrenchSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		t.Skip("french_synth.dict / official FR synth resources not available")
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

// Port of FrenchSynthesizerTest.testSynthesizeStringString
func TestFrenchSynthesizer_SynthesizeStringString(t *testing.T) {
	synth := requireFrenchInstance(t)

	// assertEquals(synth.synthesize(dummyToken("blablabla"), "blablabla").length, 0);
	forms, err := synth.Synthesize(dummyToken("blablabla"), "blablabla")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// assertEquals("[nagent]", Arrays.toString(synth.synthesize(dummyToken("nager"), "V ind pres 3 p")));
	forms, err = synth.Synthesize(dummyToken("nager"), "V ind pres 3 p")
	require.NoError(t, err)
	require.Equal(t, "[nagent]", arraysToString(forms))

	// assertEquals("[les]", Arrays.toString(synth.synthesize(dummyToken("le"), "D (m|e) (p|sp)", true)));
	forms, err = synth.SynthesizeRE(dummyToken("le"), "D (m|e) (p|sp)", true)
	require.NoError(t, err)
	require.Equal(t, "[les]", arraysToString(forms))

	// assertEquals("[marins]", Arrays.toString(synth.synthesize(dummyToken("marin"), "J [me] (p|sp)", true)));
	forms, err = synth.SynthesizeRE(dummyToken("marin"), "J [me] (p|sp)", true)
	require.NoError(t, err)
	require.Equal(t, "[marins]", arraysToString(forms))

	// assertEquals("[marins]", Arrays.toString(synth.synthesize(new AnalyzedToken("marine", "J f s", "marin"), "J [me] (p|sp)", true)));
	tag1 := "J f s"
	lemma1 := "marin"
	tok1 := languagetool.NewAnalyzedToken("marine", &tag1, &lemma1)
	forms, err = synth.SynthesizeRE(tok1, "J [me] (p|sp)", true)
	require.NoError(t, err)
	require.Equal(t, "[marins]", arraysToString(forms))

	// assertEquals("[marinés]", Arrays.toString(synth.synthesize(new AnalyzedToken("marine", "V ind pres 1 s", "mariner"), "V ppa [me] (p|sp)", true)));
	tag2 := "V ind pres 1 s"
	lemma2 := "mariner"
	tok2 := languagetool.NewAnalyzedToken("marine", &tag2, &lemma2)
	forms, err = synth.SynthesizeRE(tok2, "V ppa [me] (p|sp)", true)
	require.NoError(t, err)
	require.Equal(t, "[marinés]", arraysToString(forms))
}
