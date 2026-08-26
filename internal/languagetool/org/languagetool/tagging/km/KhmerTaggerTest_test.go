package km

// Outcome-behavior matrix for org.languagetool.tagging.km.KhmerTagger.
// Upstream has no KhmerTaggerTest.java; expectations derived from the same
// official khmer.dict + BaseTagger case-merge (TestTools.myAssert style).

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// TestKhmerTagger_DictionaryPath asserts Java KhmerTagger super path constant.
func TestKhmerTagger_DictionaryPath(t *testing.T) {
	tagger := NewKhmerTagger(nil)
	require.Equal(t, "/km/khmer.dict", KhmerDictPath)
	require.Equal(t, KhmerDictPath, tagger.GetDictionaryPath())
	require.Equal(t, "km", tagger.LocaleLanguage)
	require.True(t, tagger.TagLowercaseWithUppercase,
		"Java BaseTagger(filename, Locale) defaults tagLowercaseWithUppercase=true")
}

// TestKhmerTagger_Tagger asserts myAssert-style outcomes for Khmer surfaces
// present in official khmer.dict (readings sorted like TestTools.myAssert).
func TestKhmerTagger_Tagger(t *testing.T) {
	if DiscoverKhmerPOSDict() == "" {
		t.Skip("khmer.dict not in tree")
	}
	EnsureDefaultKhmerTagger()
	require.NotNil(t, DefaultKhmerTagger)
	require.NotNil(t, DefaultKhmerTagger.GetWordTagger())
	require.NotEmpty(t, KhmerPOSDictPath(), "real khmer.dict must load")

	cases := []struct {
		input string
		want  string
	}{
		{
			"ខ្ញុំ ជា មនុស្ស",
			"ខ្ញុំ/[ខ្ញុំ]PRO -- ជា/[ជា]VB -- មនុស្ស/[មនុស្ស]NN",
		},
		{
			"គាត់ នៅ ផ្ទះ",
			"គាត់/[គាត់]PRO|គាត់/[គាត់]PRP -- នៅ/[នៅ]IN|នៅ/[នៅ]RB|នៅ/[នៅ]VB -- ផ្ទះ/[ផ្ទះ]NN",
		},
		{
			"មាន សៀវភៅ",
			"មាន/[មាន]JJ|មាន/[មាន]NNP|មាន/[មាន]VB -- សៀវភៅ/[សៀវភៅ]NN",
		},
		{
			"ខ្មែរ កម្ពុជា",
			"ខ្មែរ/[ខ្មែរ]JJ|ខ្មែរ/[ខ្មែរ]NN -- កម្ពុជា/[កម្ពុជា]NN|កម្ពុជា/[កម្ពុជា]NNP",
		},
		{
			"និង ឬ ដែល",
			"និង/[និង]CC|និង/[និង]IN -- ឬ/[ឬ]AW|ឬ/[ឬ]CC|ឬ/[ឬ]IN -- ដែល/[ដែល]IN",
		},
		{
			"នេះ នោះ",
			"នេះ/[នេះ]DP|នេះ/[នេះ]PRP -- នោះ/[នោះ]DP|នោះ/[នោះ]JJ",
		},
		{
			"ទៅ មក",
			"ទៅ/[ទៅ]RB|ទៅ/[ទៅ]VB -- មក/[មក]RB|មក/[មក]VB",
		},
		{
			"ល្អ ធំ តូច",
			"ល្អ/[ល្អ]JJ|ល្អ/[ល្អ]RB -- ធំ/[ធំ]JJ|ធំ/[ធំ]NN|ធំ/[ធំ]NNP|ធំ/[ធំ]PRP -- តូច/[តូច]JJ|តូច/[តូច]NNP",
		},
		{
			"បាន ធ្វើ",
			"បាន/[បាន]AUX -- ធ្វើ/[ធ្វើ]VB",
		},
		{
			"អ្នក និយាយ",
			"អ្នក/[អ្នក]NN|អ្នក/[អ្នក]PRO|អ្នក/[អ្នក]PRP -- និយាយ/[និយាយ]NN|និយាយ/[និយាយ]VB",
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

// TestKhmerTagger_Dictionary opens the real khmer.dict and checks sample
// surfaces all carry non-empty POS tags (TestTools.testDictionary spirit).
func TestKhmerTagger_Dictionary(t *testing.T) {
	if DiscoverKhmerPOSDict() == "" {
		t.Skip("khmer.dict not in tree")
	}
	EnsureDefaultKhmerTagger()
	tagger := DefaultKhmerTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, KhmerDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, KhmerPOSDictPath(), "real khmer.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	// Sample surfaces present in khmer.dict (TagWordExact = exact FSA key).
	samples := []string{
		"ខ្ញុំ", "គាត់", "នាង", "យើង", "ពួកគេ", "អ្នក",
		"ជា", "នៅ", "និង", "ឬ", "ដែល", "នឹង", "មាន",
		"ទៅ", "មក", "ធ្វើ", "និយាយ", "ឃើញ", "ដឹង",
		"ផ្ទះ", "មនុស្ស", "កុមារ", "ប្រុស", "ស្រី",
		"ល្អ", "ធំ", "តូច", "ថ្មី", "ចាស់",
		"ថ្ងៃ", "ឆ្នាំ", "ពេល", "ទីក្រុង", "ប្រទេស",
		"ភាសា", "ខ្មែរ", "កម្ពុជា", "សៀវភៅ",
		"នេះ", "នោះ", "បាន", "ពី", "នៃ", "ហើយ",
	}
	for _, w := range samples {
		tw := tagger.TagWordExact(w)
		require.NotEmpty(t, tw, "dict entry missing for %q", w)
		for _, tword := range tw {
			require.NotEmpty(t, tword.PosTag, "**** Warning-equivalent: %s/%s lacks a POS tag", w, tword.Lemma)
		}
	}
}

// TestKhmerTagger_CaseMerge covers BaseTagger case-merge guards and null POS.
// Khmer script has no upper/lower case; mixed Latin still must not invent tags.
func TestKhmerTagger_CaseMerge(t *testing.T) {
	if DiscoverKhmerPOSDict() == "" {
		t.Skip("khmer.dict not in tree")
	}
	EnsureDefaultKhmerTagger()
	tagger := DefaultKhmerTagger

	// Known Khmer surface tags via TagWord (case-merge is a no-op for caseless script).
	tw := tagger.TagWord("ខ្ញុំ")
	require.NotEmpty(t, tw)
	found := false
	for _, r := range tw {
		if r.Lemma == "ខ្ញុំ" && r.PosTag == "PRO" {
			found = true
		}
	}
	require.True(t, found, "ខ្ញុំ should carry lemma/PRO: %#v", tw)

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
	EnsureDefaultKhmerTagger()
	tagger := DefaultKhmerTagger
	wt := tokenizers.NewWordTokenizer()
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
