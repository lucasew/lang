package en_test

// Twin of org.languagetool.synthesis.en.EnglishSynthesizerTest using official
// EnglishSynthesizer.INSTANCE resources (english_synth.dict, english_tags.txt,
// added.txt, removed.txt, do-not-synthesize.txt, en.sor) and AvsAnRule for +DT/+INDT.

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	rulesen "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/en"
	"github.com/stretchr/testify/require"
)

func requireEnglishInstance(t *testing.T) *en.EnglishSynthesizer {
	t.Helper()
	s := en.Instance()
	if s == nil {
		t.Skip("english_synth.dict / official EN synth resources not available")
	}
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Same(t, s, en.INSTANCE)
	// Java: private final AvsAnRule aVsAnRule = new AvsAnRule(...);
	// rules/en init also sets DefaultSuggestAorAn; set field explicitly for twin clarity.
	if s.SuggestAorAn == nil {
		s.SuggestAorAn = rulesen.NewAvsAnRule(nil).SuggestAorAn
	}
	return s
}

func dummyToken(tokenStr string) *languagetool.AnalyzedToken {
	// Java: new AnalyzedToken(tokenStr, tokenStr, tokenStr)
	tag := tokenStr
	lem := tokenStr
	return languagetool.NewAnalyzedToken(tokenStr, &tag, &lem)
}

func dummyTokenLemma(tokenStr, tokenLemma string) *languagetool.AnalyzedToken {
	// Java: new AnalyzedToken(tokenStr, tokenStr, tokenLemma)
	tag := tokenStr
	lem := tokenLemma
	return languagetool.NewAnalyzedToken(tokenStr, &tag, &lem)
}

// arraysToString ports Java Arrays.toString(String[]) for twin assertions.
func arraysToString(forms []string) string {
	if forms == nil {
		return "[]"
	}
	return "[" + strings.Join(forms, ", ") + "]"
}

// Port of EnglishSynthesizerTest.testSynthesizeStringString
func TestEnglishSynthesizer_SynthesizeStringString(t *testing.T) {
	synth := requireEnglishInstance(t)

	// assertEquals(synth.synthesize(dummyToken("blablabla"), "blablabla").length, 0);
	forms, err := synth.Synthesize(dummyToken("blablabla"), "blablabla")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// assertEquals("[was, were]", Arrays.toString(synth.synthesize(dummyToken("be"), "VBD")));
	forms, err = synth.Synthesize(dummyToken("be"), "VBD")
	require.NoError(t, err)
	require.Equal(t, "[was, were]", arraysToString(forms))

	// assertEquals("[presidents]", Arrays.toString(synth.synthesize(dummyToken("president"), "NNS")));
	forms, err = synth.Synthesize(dummyToken("president"), "NNS")
	require.NoError(t, err)
	require.Equal(t, "[presidents]", arraysToString(forms))

	// assertEquals("[tested]", Arrays.toString(synth.synthesize(dummyToken("test"), "VBD")));
	forms, err = synth.Synthesize(dummyToken("test"), "VBD")
	require.NoError(t, err)
	require.Equal(t, "[tested]", arraysToString(forms))

	// assertEquals("[tested]", Arrays.toString(synth.synthesize(dummyToken("test"), "VBD", false)));
	forms, err = synth.SynthesizeRE(dummyToken("test"), "VBD", false)
	require.NoError(t, err)
	require.Equal(t, "[tested]", arraysToString(forms))

	// with regular expressions
	// assertEquals("[tested]", Arrays.toString(synth.synthesize(dummyToken("test"), "VBD", true)));
	forms, err = synth.SynthesizeRE(dummyToken("test"), "VBD", true)
	require.NoError(t, err)
	require.Equal(t, "[tested]", arraysToString(forms))

	// assertEquals("[tested, testing]", Arrays.toString(synth.synthesize(dummyToken("test"), "VBD|VBG", true)));
	forms, err = synth.SynthesizeRE(dummyToken("test"), "VBD|VBG", true)
	require.NoError(t, err)
	require.Equal(t, "[tested, testing]", arraysToString(forms))

	// with special indefinite article
	// assertEquals("[a university, the university]", Arrays.toString(synth.synthesize(dummyToken("university"), "+DT", false)));
	forms, err = synth.SynthesizeRE(dummyToken("university"), en.AddDeterminer, false)
	require.NoError(t, err)
	require.Equal(t, "[a university, the university]", arraysToString(forms))

	// assertEquals("[an hour, the hour]", Arrays.toString(synth.synthesize(dummyToken("hour"), "+DT", false)));
	forms, err = synth.SynthesizeRE(dummyToken("hour"), en.AddDeterminer, false)
	require.NoError(t, err)
	require.Equal(t, "[an hour, the hour]", arraysToString(forms))

	// assertEquals("[an hour]", Arrays.toString(synth.synthesize(dummyToken("hour"), "+INDT", false)));
	forms, err = synth.SynthesizeRE(dummyToken("hour"), en.AddIndDeterminer, false)
	require.NoError(t, err)
	require.Equal(t, "[an hour]", arraysToString(forms))

	// indefinite article and other changes...
	// assertEquals("[an hour]", Arrays.toString(synth.synthesize(dummyToken("hours", "hour"), "NN\\+INDT", true)));
	forms, err = synth.SynthesizeRE(dummyTokenLemma("hours", "hour"), `NN\+INDT`, true)
	require.NoError(t, err)
	require.Equal(t, "[an hour]", arraysToString(forms))

	// assertEquals("[a hexagon]", Arrays.toString(synth.synthesize(dummyToken("hexagon"), "NN|NN:.*\\+INDT", true)));
	forms, err = synth.SynthesizeRE(dummyToken("hexagon"), `NN|NN:.*\+INDT`, true)
	require.NoError(t, err)
	require.Equal(t, "[a hexagon]", arraysToString(forms))

	// assertEquals("[the hour]", Arrays.toString(synth.synthesize(dummyToken("hours", "hour"), "NN\\+DT", true)));
	forms, err = synth.SynthesizeRE(dummyTokenLemma("hours", "hour"), `NN\+DT`, true)
	require.NoError(t, err)
	require.Equal(t, "[the hour]", arraysToString(forms))

	// from added.txt:
	// assertEquals("[absolutized]", Arrays.toString(synth.synthesize(dummyToken("absolutize"), "VBD", false)));
	forms, err = synth.SynthesizeRE(dummyToken("absolutize"), "VBD", false)
	require.NoError(t, err)
	require.Equal(t, "[absolutized]", arraysToString(forms))

	// assertEquals("[absolutized]", Arrays.toString(synth.synthesize(dummyToken("absolutize"), "VB[XD]", true)));
	forms, err = synth.SynthesizeRE(dummyToken("absolutize"), "VB[XD]", true)
	require.NoError(t, err)
	require.Equal(t, "[absolutized]", arraysToString(forms))

	// from removed.txt:
	// assertEquals("[]", Arrays.toString(synth.synthesize(dummyToken("Christmas"), "VBZ", false)));
	forms, err = synth.SynthesizeRE(dummyToken("Christmas"), "VBZ", false)
	require.NoError(t, err)
	require.Equal(t, "[]", arraysToString(forms))

	// assertEquals("[twelve]", Arrays.toString(synth.synthesize(dummyToken("12"), "_spell_number_", false)));
	forms, err = synth.SynthesizeRE(dummyToken("12"), synthesis.SpellNumberTag, false)
	require.NoError(t, err)
	require.Equal(t, "[twelve]", arraysToString(forms))

	// assertEquals("[one thousand two hundred forty-three]", Arrays.toString(synth.synthesize(dummyToken("1243"), "_spell_number_", false)));
	forms, err = synth.SynthesizeRE(dummyToken("1243"), synthesis.SpellNumberTag, false)
	require.NoError(t, err)
	require.Equal(t, "[one thousand two hundred forty-three]", arraysToString(forms))

	// assertEquals("[twelve]", Arrays.toString(synth.synthesize(dummyToken("12"), "_spell_number_", true)));
	forms, err = synth.SynthesizeRE(dummyToken("12"), synthesis.SpellNumberTag, true)
	require.NoError(t, err)
	require.Equal(t, "[twelve]", arraysToString(forms))

	// assertEquals("[I]", Arrays.toString(synth.synthesize(dummyToken("myself", "I"), "PRP_S1S", true)));
	forms, err = synth.SynthesizeRE(dummyTokenLemma("myself", "I"), "PRP_S1S", true)
	require.NoError(t, err)
	require.Equal(t, "[I]", arraysToString(forms))

	// assertEquals("[my]", Arrays.toString(synth.synthesize(dummyToken("myself", "I"), "PRP\\$_A1S", true)));
	forms, err = synth.SynthesizeRE(dummyTokenLemma("myself", "I"), `PRP\$_A1S`, true)
	require.NoError(t, err)
	require.Equal(t, "[my]", arraysToString(forms))

	// assertEquals("[mixed]", Arrays.toString(synth.synthesize(dummyToken("mix"), "VBD")));
	forms, err = synth.Synthesize(dummyToken("mix"), "VBD")
	require.NoError(t, err)
	require.Equal(t, "[mixed]", arraysToString(forms))

	// assertEquals("[mixed]", Arrays.toString(synth.synthesize(dummyToken("mix"), "VBN")));
	forms, err = synth.Synthesize(dummyToken("mix"), "VBN")
	require.NoError(t, err)
	require.Equal(t, "[mixed]", arraysToString(forms))
}
