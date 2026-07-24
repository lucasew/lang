package ml

// Outcome-behavior matrix for org.languagetool.tagging.ml.MalayalamTagger.
// Upstream has no MalayalamTaggerTest.java; expectations derived from the same
// official malayalam.dict + BaseTagger case-merge (TestTools.myAssert style).

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	mltok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ml"
	"github.com/stretchr/testify/require"
)

// TestMalayalamTagger_DictionaryPath asserts Java MalayalamTagger super path constant.
func TestMalayalamTagger_DictionaryPath(t *testing.T) {
	tagger := NewMalayalamTagger(nil)
	require.Equal(t, "/ml/malayalam.dict", MalayalamDictPath)
	require.Equal(t, MalayalamDictPath, tagger.GetDictionaryPath())
	require.Equal(t, "ml", tagger.LocaleLanguage)
	require.True(t, tagger.TagLowercaseWithUppercase,
		"Java BaseTagger(filename, Locale) defaults tagLowercaseWithUppercase=true")
}

// TestMalayalamTagger_Tagger asserts myAssert-style outcomes for Malayalam surfaces
// present in official malayalam.dict (readings sorted like TestTools.myAssert).
func TestMalayalamTagger_Tagger(t *testing.T) {
	if DiscoverMalayalamPOSDict() == "" {
		t.Skip("malayalam.dict not in tree")
	}
	EnsureDefaultMalayalamTagger()
	require.NotNil(t, DefaultMalayalamTagger)
	require.NotNil(t, DefaultMalayalamTagger.GetWordTagger())
	require.NotEmpty(t, MalayalamPOSDictPath(), "real malayalam.dict must load")

	cases := []struct {
		input string
		want  string
	}{
		{
			"ആണ് ആരംഭിക്കുന്നു",
			"ആണ്/[ആണ്]VBB -- ആരംഭിക്കുന്നു/[ആരംഭിക്ക്]VBP",
		},
		{
			"പഴയ രൂപം",
			"പഴയ/[പഴയ]JJ -- രൂപം/[രൂപം]NN1",
		},
		{
			"ഭാഗം സമൂഹം",
			"ഭാഗം/[ഭാഗം]NN1 -- സമൂഹം/[സമൂഹം]NN1",
		},
		{
			"നൂതന ഉത്തമമായ",
			"നൂതന/[നൂതന]JJ -- ഉത്തമമായ/[ഉത്തമ]JJ",
		},
		{
			"എന്ന് മാത്രം",
			"എന്ന്/[എന്ന്]AVY -- മാത്രം/[മാത്രം]AVY",
		},
		{
			"പറഞ്ഞ് വിടാതെ",
			"പറഞ്ഞ്/[പറഞ്ഞ്]VP -- വിടാതെ/[വിട്]VPN",
		},
		{
			"കഥകള്‍ രചനകള്‍",
			"കഥകള്‍/[കഥ]NNS -- രചനകള്‍/[രചന]NNS",
		},
		{
			"തുടങ്ങി തുടങ്ങിയ",
			"തുടങ്ങി/[തുടങ്ങ്]VBD -- തുടങ്ങിയ/[തുടങ്ങ്]JJ",
		},
		{
			"ഉം ആണ്",
			"ഉം/[ഉം]CJC -- ആണ്/[ആണ്]VBB",
		},
		{
			"ശ്രദ്ധേയം ശ്രദ്ധേയമായ",
			"ശ്രദ്ധേയം/[ശ്രദ്ധേയം]RB -- ശ്രദ്ധേയമായ/[ശ്രദ്ധേയം]JJ",
		},
		{
			"വടക്കാഞ്ചേരി ദേശമംഗലം",
			"വടക്കാഞ്ചേരി/[വടക്കാഞ്ചേരി]NNP -- ദേശമംഗലം/[ദേശമംഗലം]NPR",
		},
		// Unknown surface → null POS reading (BaseTagger empty → invent null token).
		{
			"xyzzyqqq",
			"xyzzyqqq/[null]null",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// TestMalayalamTagger_Dictionary opens the real malayalam.dict and checks sample
// surfaces all carry non-empty POS tags (TestTools.testDictionary spirit).
func TestMalayalamTagger_Dictionary(t *testing.T) {
	if DiscoverMalayalamPOSDict() == "" {
		t.Skip("malayalam.dict not in tree")
	}
	EnsureDefaultMalayalamTagger()
	tagger := DefaultMalayalamTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, MalayalamDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, MalayalamPOSDictPath(), "real malayalam.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	// Sample surfaces present in malayalam.dict (TagWordExact = exact FSA key).
	// Official dict is small (~87 entries); samples taken from that set.
	samples := []string{
		"ആണ്", "ആരംഭിക്കുന്നു", "ഉം", "എന്ന്", "പഴയ", "രൂപം",
		"ഭാഗം", "സമൂഹം", "നൂതന", "പറഞ്ഞ്", "മാത്രം", "കഥകള്‍",
		"രചനകള്‍", "തുടങ്ങി", "തുടങ്ങിയ", "ശ്രദ്ധേയം", "ശ്രദ്ധേയമായ",
		"വടക്കാഞ്ചേരി", "ദേശമംഗലം", "ഉത്തമമായ", "വിടാതെ", "സാധാരണം",
		"പ്രസ്ഥാനം", "ആലയം", "ഫലിതം", "സമാനം", "സാധ്യമല്ല", "കാണുന്നില്ല",
		"അതിനെ", "അതിന്", "അതില്‍", "പഴക്കം", "പുരോഗതി", "വിശേഷം",
	}
	for _, w := range samples {
		tw := tagger.TagWordExact(w)
		require.NotEmpty(t, tw, "dict entry missing for %q", w)
		for _, tword := range tw {
			require.NotEmpty(t, tword.PosTag, "**** Warning-equivalent: %s/%s lacks a POS tag", w, tword.Lemma)
		}
	}
}

// TestMalayalamTagger_CaseMerge covers BaseTagger case-merge guards and null POS.
// Malayalam script has no upper/lower case; mixed Latin still must not invent tags.
func TestMalayalamTagger_CaseMerge(t *testing.T) {
	if DiscoverMalayalamPOSDict() == "" {
		t.Skip("malayalam.dict not in tree")
	}
	EnsureDefaultMalayalamTagger()
	tagger := DefaultMalayalamTagger

	// Known Malayalam surface tags via TagWord (case-merge is a no-op for caseless script).
	tw := tagger.TagWord("ആണ്")
	require.NotEmpty(t, tw)
	found := false
	for _, r := range tw {
		if r.Lemma == "ആണ്" && r.PosTag == "VBB" {
			found = true
		}
	}
	require.True(t, found, "ആണ് should carry lemma/VBB: %#v", tw)

	// Mixed case Latin must not merge lower tags (BaseTagger isMixedCase guard).
	// "BaHay" is mixed → only exact surface lookup (empty) → no lower merge.
	mixed := tagger.TagWord("BaHay")
	require.Empty(t, mixed, "mixed-case should not receive lower tags")

	// Unknown → Tag() null POS.
	out := tagger.Tag([]string{"xyzzyqqq"})
	require.Len(t, out, 1)
	readings := out[0].GetReadings()
	require.Len(t, readings, 1)
	require.Nil(t, readings[0].GetPOSTag())
}

// myAssertTagger ports Java TestTools.myAssert(input, expected, tokenizer, tagger):
// tokenize → drop non-word tokens → tag → sorted readings joined by " -- ".
func myAssertTagger(input string) string {
	EnsureDefaultMalayalamTagger()
	tagger := DefaultMalayalamTagger
	// Java Malayalam language uses MalayalamWordTokenizer (not core WordTokenizer);
	// ZWJ (U+200D) in dict surfaces must not be split as a delimiter.
	wt := mltok.NewMalayalamWordTokenizer()
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
