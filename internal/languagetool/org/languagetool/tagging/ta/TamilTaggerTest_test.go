package ta

// Outcome-behavior matrix for org.languagetool.language.tagging.TamilTagger.
// Upstream has no TamilTaggerTest.java; expectations derived from the same
// official tamil.dict + BaseTagger case-merge (TestTools.myAssert style).

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// TestTamilTagger_DictionaryPath asserts Java TamilTagger super path constant.
func TestTamilTagger_DictionaryPath(t *testing.T) {
	tagger := NewTamilTagger(nil)
	require.Equal(t, "/ta/tamil.dict", TamilDictPath)
	require.Equal(t, TamilDictPath, tagger.GetDictionaryPath())
	require.Equal(t, "ta", tagger.LocaleLanguage)
	require.True(t, tagger.TagLowercaseWithUppercase,
		"Java BaseTagger(filename, Locale) defaults tagLowercaseWithUppercase=true")
}

// TestTamilTagger_Tagger asserts myAssert-style outcomes for Tamil surfaces
// present in official tamil.dict (readings sorted like TestTools.myAssert).
func TestTamilTagger_Tagger(t *testing.T) {
	if DiscoverTamilPOSDict() == "" {
		t.Skip("tamil.dict not in tree")
	}
	EnsureDefaultTamilTagger()
	require.NotNil(t, DefaultTamilTagger)
	require.NotNil(t, DefaultTamilTagger.GetWordTagger())
	require.NotEmpty(t, TamilPOSDictPath(), "real tamil.dict must load")

	cases := []struct {
		input string
		want  string
	}{
		{
			"நான் அவன் அவள்",
			"நான்/[நான்]NNPU -- அவன்/[அவன்]NNPUM -- அவள்/[அவள்]NNPUF",
		},
		{
			"நீங்கள் அவர்கள்",
			"நீங்கள்/[நீ]NNSPU -- அவர்கள்/[அவர்]NNSPU",
		},
		{
			"செய்ய செய்து வா",
			"செய்ய/[செய்]VAN -- செய்து/[செய்]VP -- வா/[வா]VB",
		},
		{
			"வர அடிக்க அழ",
			"வர/[வா]VAN -- அடிக்க/[அடி]VAN -- அழ/[அழு]VAN",
		},
		{
			"நல்ல அழகிய",
			"நல்ல/[நல்ல]ADJ -- அழகிய/[அழகு]ADJ",
		},
		{
			"படித்த படிக்காத படிக்கா",
			"படித்த/[படி]RP -- படிக்காத/[படி]RPN -- படிக்கா/[படி]RPNN",
		},
		{
			"தந்த தராத தரா",
			"தந்த/[தா]RP -- தராத/[தா]RPN -- தரா/[தா]RPNN",
		},
		{
			"யானை யானைகள் வானூர்தி",
			"யானை/[யானை]NNA -- யானைகள்/[யானை]NNSA -- வானூர்தி/[வானூர்தி]NNA",
		},
		{
			"எழுத்தாளர் எழுத்தாளர்கள்",
			"எழுத்தாளர்/[எழுத்தாளர்]NNU -- எழுத்தாளர்கள்/[எழுத்தாளர்]NNSU",
		},
		{
			"அப்பா அண்ணன் அக்கா அழகி",
			"அப்பா/[அப்பா]NNUM|அப்பா/[அப்பு]RPNN -- அண்ணன்/[அண்ணன்]NNUM -- அக்கா/[அக்கா]NNUF -- அழகி/[அழகி]NNUF",
		},
		{
			"அவனை அவனுக்கு அவளை அவளுக்கு",
			"அவனை/[அவன்]NNPUM-S -- அவனுக்கு/[அவன்]NNPUM-F -- அவளை/[அவள்]NNPUF-S -- அவளுக்கு/[அவள்]NNPUF-F",
		},
		{
			"அவர் அவரை அவருக்கு",
			"அவர்/[அவர்]NNPU -- அவரை/[அவர்]NNPU-S -- அவருக்கு/[அவர்]NNPU-F",
		},
		{
			"அது அதை அதற்கு",
			"அது/[அது]NNPA -- அதை/[அது]NNPA-S -- அதற்கு/[அது]NNPA-F",
		},
		{
			"வழி செல்",
			"வழி/[வழி]NNA|வழி/[வழி]VB -- செல்/[செல்]NNA|செல்/[செல்]VB",
		},
		{
			"இது நாள் வேலை",
			"இது/[இது]NNPA -- நாள்/[நாள்]NNA -- வேலை/[வேலை]NNA",
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

// TestTamilTagger_Dictionary opens the real tamil.dict and checks sample
// surfaces all carry non-empty POS tags (TestTools.testDictionary spirit).
func TestTamilTagger_Dictionary(t *testing.T) {
	if DiscoverTamilPOSDict() == "" {
		t.Skip("tamil.dict not in tree")
	}
	EnsureDefaultTamilTagger()
	tagger := DefaultTamilTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, TamilDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, TamilPOSDictPath(), "real tamil.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	// Sample surfaces present in tamil.dict (TagWordExact = exact FSA key).
	samples := []string{
		"நான்", "அவன்", "அவள்", "நீ", "நாங்கள்", "நீங்கள்", "அவர்கள்",
		"செய்ய", "செய்து", "வா", "வர", "அடிக்க", "அழ", "செய்",
		"நல்ல", "அழகிய", "படித்த", "படிக்காத", "படிக்கா",
		"தந்த", "தராத", "தரா", "யானை", "யானைகள்", "வானூர்தி",
		"எழுத்தாளர்", "எழுத்தாளர்கள்", "அப்பா", "அண்ணன்", "அக்கா", "அழகி",
		"அவனை", "அவனுக்கு", "அவளை", "அவளுக்கு",
		"அவர்", "அவரை", "அவருக்கு", "அது", "அதை", "அதற்கு",
		"வழி", "செல்", "இது", "நாள்", "வேலை", "உன்னை", "எனக்கு",
		"போய்", "பற்றி", "நிறைய", "ஓடி", "அடித்து",
	}
	for _, w := range samples {
		tw := tagger.TagWordExact(w)
		require.NotEmpty(t, tw, "dict entry missing for %q", w)
		for _, tword := range tw {
			require.NotEmpty(t, tword.PosTag, "**** Warning-equivalent: %s/%s lacks a POS tag", w, tword.Lemma)
		}
	}
}

// TestTamilTagger_CaseMerge covers BaseTagger case-merge guards and null POS.
// Tamil script has no upper/lower case; mixed Latin still must not invent tags.
func TestTamilTagger_CaseMerge(t *testing.T) {
	if DiscoverTamilPOSDict() == "" {
		t.Skip("tamil.dict not in tree")
	}
	EnsureDefaultTamilTagger()
	tagger := DefaultTamilTagger

	// Known Tamil surface tags via TagWord (case-merge is a no-op for caseless script).
	tw := tagger.TagWord("நான்")
	require.NotEmpty(t, tw)
	found := false
	for _, r := range tw {
		if r.Lemma == "நான்" && r.PosTag == "NNPU" {
			found = true
		}
	}
	require.True(t, found, "நான் should carry lemma/NNPU: %#v", tw)

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
	EnsureDefaultTamilTagger()
	tagger := DefaultTamilTagger
	// Java Tamil language uses default WordTokenizer (no language-specific override).
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
