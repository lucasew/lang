package en

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
	"github.com/stretchr/testify/require"
)

// Twin of org.languagetool.tagging.en.EnglishTaggerTest#testTagger
// (Java TestTools.myAssert with EnglishWordTokenizer + EnglishTagger).
func TestEnglishTagger_Tagger(t *testing.T) {
	if DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree")
	}
	EnsureDefaultEnglishTagger()
	require.NotNil(t, DefaultEnglishTagger)
	require.NotNil(t, DefaultEnglishTagger.GetWordTagger())

	// Java EnglishTaggerTest.testTagger expected strings (readings sorted in TestTools).
	cases := []struct {
		input string
		want  string
	}{
		{
			"This is a big house.",
			"This/[this]DT|This/[this]PDT -- is/[be]VBZ -- a/[a]DT -- big/[big]JJ|big/[big]RB -- house/[house]NN|house/[house]VB|house/[house]VBP",
		},
		{
			"Marketing do a lot of trouble.",
			"Marketing/[market]VBG|Marketing/[marketing]NN:U -- do/[do]VB|do/[do]VBP -- a/[a]DT -- lot/[lot]NN -- of/[of]IN -- trouble/[trouble]NN:UN|trouble/[trouble]VB|trouble/[trouble]VBP",
		},
		{
			"Manager use his laptop every day.",
			"Manager/[manager]NN -- use/[use]NN:UN|use/[use]VB|use/[use]VBP -- his/[he]PRP$_A3SM|his/[he]PRP$_P3SM|his/[hi]NNS|his/[his]PRP$ -- laptop/[laptop]NN -- every/[every]DT -- day/[day]NN:UN",
		},
		{
			"This is a bigger house.",
			"This/[this]DT|This/[this]PDT -- is/[be]VBZ -- a/[a]DT -- bigger/[big]JJR -- house/[house]NN|house/[house]VB|house/[house]VBP",
		},
		{
			"He doesn't believe me.",
			"He/[he]PRP|He/[he]PRP_S3SM -- does/[do]VBZ|does/[doe]NNS -- n't/[not]RB -- believe/[believe]VB|believe/[believe]VBP -- me/[I]PRP|me/[I]PRP_O1S",
		},
		{
			"It has become difficult.",
			"It/[it]PRP|It/[it]PRP_O3SN|It/[it]PRP_S3SN -- has/[have]VBZ -- become/[become]VB|become/[become]VBN|become/[become]VBP -- difficult/[difficult]JJ",
		},
		{
			"You haven't.",
			"You/[you]PRP|You/[you]PRP_O2P|You/[you]PRP_O2S|You/[you]PRP_S2P|You/[you]PRP_S2S -- have/[have]NN|have/[have]VB|have/[have]VBP -- n't/[not]RB",
		},
		{
			// Typographic apostrophe: Java rewrites surface to typewriter ' in readings.
			"You haven’t.",
			"You/[you]PRP|You/[you]PRP_O2P|You/[you]PRP_O2S|You/[you]PRP_S2P|You/[you]PRP_S2S -- have/[have]NN|have/[have]VB|have/[have]VBP -- n't/[not]RB",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// Twin of EnglishTaggerTest.testLemma — real english.dict via INSTANCE / DefaultEnglishTagger.
func TestEnglishTagger_Lemma(t *testing.T) {
	if DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree")
	}
	EnsureDefaultEnglishTagger()
	// Java: EnglishTagger.INSTANCE.tag(words)
	got := DefaultEnglishTagger.Tag([]string{"Trump", "works"})
	require.Len(t, got, 2)
	require.Len(t, got[0].GetReadings(), 4)
	require.Len(t, got[1].GetReadings(), 2)

	require.Equal(t, "Trump", lemmaOf(got[0].GetReadings()[0]))
	require.Equal(t, "trump", lemmaOf(got[0].GetReadings()[1]))
	require.Equal(t, "trump", lemmaOf(got[0].GetReadings()[2]))
	require.Equal(t, "trump", lemmaOf(got[0].GetReadings()[3]))

	require.Equal(t, "work", lemmaOf(got[1].GetReadings()[0]))
	require.Equal(t, "work", lemmaOf(got[1].GetReadings()[1]))
}

// Twin of EnglishTaggerTest.testDictionary (Java TestTools.testDictionary).
// Java walks every Morfologik WordData and only warns on empty POS; no assert fail.
// Full FSA DictionaryLookup iteration is not yet ported; this opens the real
// english.dict and checks sample surfaces all carry non-empty POS tags.
func TestEnglishTagger_Dictionary(t *testing.T) {
	if DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree")
	}
	EnsureDefaultEnglishTagger()
	tagger := DefaultEnglishTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, EnglishDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, EnglishPOSDictPath(), "real english.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	// Sample of surfaces present in Java testTagger / lexicon — each must have POS.
	samples := []string{"house", "this", "be", "big", "market", "manager", "have", "work", "trump"}
	for _, w := range samples {
		tw := tagger.TagWordExact(w)
		require.NotEmpty(t, tw, "dict entry missing for %q", w)
		for _, tword := range tw {
			require.NotEmpty(t, tword.PosTag, "**** Warning-equivalent: %s/%s lacks a POS tag", w, tword.Lemma)
		}
	}
}

// myAssertTagger ports Java TestTools.myAssert(input, expected, tokenizer, tagger):
// tokenize → drop non-word tokens → tag → sorted readings joined by " -- ".
func myAssertTagger(input string) string {
	EnsureDefaultEnglishTagger()
	tagger := DefaultEnglishTagger
	wt := entok.NewEnglishWordTokenizer()
	tokens := wt.Tokenize(input)
	var noWS []string
	for _, tok := range tokens {
		if testToolsIsWord(tok) {
			noWS = append(noWS, tok)
		}
	}
	output := tagger.Tag(noWS)
	var parts []string
	for _, atr := range output {
		parts = append(parts, strings.Join(testToolsGetAsStrings(atr), "|"))
	}
	return strings.Join(parts, " -- ")
}

// testToolsIsWord ports TestTools.isWord: any letter or digit → word token.
func testToolsIsWord(token string) bool {
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// testToolsGetAsStrings ports TestTools.getAsStrings (sorted).
func testToolsGetAsStrings(atr *languagetool.AnalyzedTokenReadings) []string {
	if atr == nil {
		return nil
	}
	var readings []string
	for _, r := range atr.GetReadings() {
		if r != nil {
			readings = append(readings, testToolsGetAsString(r))
		}
	}
	sort.Strings(readings)
	return readings
}

// testToolsGetAsString ports TestTools.getAsString: token/[lemma]POS with null literals.
func testToolsGetAsString(tok *languagetool.AnalyzedToken) string {
	lemma, pos := "null", "null"
	if tok.GetLemma() != nil {
		lemma = *tok.GetLemma()
	}
	if tok.GetPOSTag() != nil {
		pos = *tok.GetPOSTag()
	}
	return tok.GetToken() + "/[" + lemma + "]" + pos
}

func lemmaOf(tok *languagetool.AnalyzedToken) string {
	if tok == nil || tok.GetLemma() == nil {
		return ""
	}
	return *tok.GetLemma()
}
