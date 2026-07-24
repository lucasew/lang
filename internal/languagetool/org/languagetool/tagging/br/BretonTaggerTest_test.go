package br

// Outcome-behavior matrix for org.languagetool.tagging.br.BretonTagger.
// Upstream has no BretonTaggerTest.java; expectations derived from the same
// official breton.dict + BretonTagger.tag control flow (TestTools.myAssert style).

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	brtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/br"
	"github.com/stretchr/testify/require"
)

// TestBretonTagger_DictionaryPath asserts Java BretonTagger super path constant.
func TestBretonTagger_DictionaryPath(t *testing.T) {
	tagger := NewBretonTagger(nil)
	require.Equal(t, "/br/breton.dict", BretonDictPath)
	require.Equal(t, BretonDictPath, BretonTaggerDictPath)
	require.Equal(t, BretonDictPath, tagger.GetDictionaryPath())
	require.Equal(t, "br", tagger.LocaleLanguage)
	require.True(t, tagger.TagLowercaseWithUppercase,
		"Java BaseTagger(filename, Locale) defaults tagLowercaseWithUppercase=true")
}

// TestBretonTagger_Tagger asserts myAssert-style outcomes for Breton surfaces
// present in official breton.dict (readings sorted like TestTools.myAssert).
func TestBretonTagger_Tagger(t *testing.T) {
	if DiscoverBretonPOSDict() == "" {
		t.Skip("breton.dict not in tree")
	}
	EnsureDefaultBretonTagger()
	require.NotNil(t, DefaultBretonTagger)
	require.NotNil(t, DefaultBretonTagger.GetWordTagger())
	require.NotEmpty(t, BretonPOSDictPath(), "real breton.dict must load")

	cases := []struct {
		input string
		want  string
	}{
		// Single-reading known surfaces.
		{
			"eo",
			"eo/[bezañ]V pres 3 s",
		},
		{
			"ti",
			"ti/[ti]N m s",
		},
		{
			"kafe",
			"kafe/[kafe]N m s",
		},
		{
			"kêr",
			"kêr/[kêr]N f s",
		},
		{
			"tud",
			"tud/[den]N m p t",
		},
		{
			"studierien",
			"studierien/[studier]N m p t",
		},
		{
			"pinvidik",
			"pinvidik/[pinvidik]J",
		},
		{
			"dont",
			"dont/[dont]V inf",
		},
		{
			"mont",
			"mont/[mont]V inf",
		},
		{
			"bezañ",
			"bezañ/[bezañ]V inf",
		},
		{
			"amañ",
			"amañ/[amañ]A",
		},
		{
			// Title case: exact Breizh + lower "breizh"→preizh (no isMixedCase skip).
			"Breizh",
			"Breizh/[Breizh]Z e s top|Breizh/[preizh]N m s M:1:1a:",
		},
		// Multi-word sentence fragment.
		{
			"eo mat",
			"eo/[bezañ]V pres 3 s -- mat/[mat]J",
		},
		// Demonstrative suffix strip: forms not in dict as whole, only after -mañ/-se/-hont.
		{
			"ti-mañ",
			"ti-mañ/[ti]N m s",
		},
		{
			"ti-se",
			"ti-se/[ti]N m s",
		},
		{
			"ti-hont",
			"ti-hont/[ti]N m s",
		},
		{
			"kêr-mañ",
			"kêr-mañ/[kêr]N f s",
		},
		{
			"kêr-se",
			"kêr-se/[kêr]N f s",
		},
		{
			"kêr-hont",
			"kêr-hont/[kêr]N f s",
		},
		{
			"heol-mañ",
			"heol-mañ/[heol]N m s",
		},
		{
			"mor-se",
			"mor-se/[mor]N m s|mor-se/[morañ]V impe 2 s|mor-se/[morañ]V pres 3 s",
		},
		// Case-insensitive suffix match (pattern (?iu)).
		{
			"ti-MAÑ",
			"ti-MAÑ/[ti]N m s",
		},
		// Unknown surface → null POS reading.
		{
			"xyzzyqqq",
			"xyzzyqqq/[null]null",
		},
		// Stem shorter than pattern (..+) → no strip → null.
		{
			"a-mañ",
			"a-mañ/[null]null",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// TestBretonTagger_Dictionary opens the real breton.dict and checks sample
// surfaces all carry non-empty POS tags (TestTools.testDictionary spirit).
func TestBretonTagger_Dictionary(t *testing.T) {
	if DiscoverBretonPOSDict() == "" {
		t.Skip("breton.dict not in tree")
	}
	EnsureDefaultBretonTagger()
	tagger := DefaultBretonTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, BretonDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, BretonPOSDictPath(), "real breton.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	// Sample surfaces present in breton.dict (TagWordExact = exact FSA key).
	samples := []string{
		"eo", "oa", "zo", "ti", "kafe", "bara", "kêr", "tud", "den", "tad", "mamm",
		"studierien", "pinvidik", "binvidik", "dont", "mont", "bezañ", "ober",
		"gwelout", "komz", "skrivañ", "lenn", "debriñ", "evañ", "kanañ", "labourat",
		"mor", "menez", "heol", "avel", "amzer", "bloaz", "deiz", "noz", "mintin",
		"gwenn", "bras", "bihan", "hir", "berr", "yen", "tomm", "fall", "brav",
		"amañ", "aze", "ahont", "bremañ", "ket", "mat", "gant", "war", "evit",
		"Breizh", "Gelted", "yezh", "bro", "skol", "hent", "bugel", "bugale",
	}
	for _, w := range samples {
		tw := tagger.TagWordExact(w)
		require.NotEmpty(t, tw, "dict entry missing for %q", w)
		for _, tword := range tw {
			require.NotEmpty(t, tword.PosTag, "**** Warning-equivalent: %s/%s lacks a POS tag", w, tword.Lemma)
		}
	}
}

// TestBretonTagger_SuffixAndLength covers suffix strip retry, length>50, case merge quirks.
func TestBretonTagger_SuffixAndLength(t *testing.T) {
	if DiscoverBretonPOSDict() == "" {
		t.Skip("breton.dict not in tree")
	}
	EnsureDefaultBretonTagger()
	tagger := DefaultBretonTagger

	// length > 50 (Java String.length = UTF-16) → null POS, no dict probe.
	long := strings.Repeat("a", 51)
	out := tagger.Tag([]string{long})
	require.Len(t, out, 1)
	readings := out[0].GetReadings()
	require.Len(t, readings, 1)
	require.Nil(t, readings[0].GetPOSTag())
	require.Equal(t, long, readings[0].GetToken())

	// Exactly 50 is still probed (not > 50).
	exact50 := strings.Repeat("a", 50)
	out50 := tagger.Tag([]string{exact50})
	require.Len(t, out50, 1)
	// unknown 50-char word → null POS after empty dict
	require.Nil(t, out50[0].GetReadings()[0].GetPOSTag())

	// Suffix strip: surface remains the original form.
	outSuf := tagger.Tag([]string{"ti-mañ"})
	require.Len(t, outSuf, 1)
	r := outSuf[0].GetReadings()
	require.NotEmpty(t, r)
	require.Equal(t, "ti-mañ", r[0].GetToken())
	require.NotNil(t, r[0].GetPOSTag())
	require.Equal(t, "N m s", *r[0].GetPOSTag())
	require.NotNil(t, r[0].GetLemma())
	require.Equal(t, "ti", *r[0].GetLemma())

	// BretonTagger differs from BaseTagger: NO isMixedCase skip for lower merge.
	// "BaRa" is mixed; exact empty; lower "bara" hits → tags attached with surface BaRa.
	mixed := tagger.Tag([]string{"BaRa"})
	require.Len(t, mixed, 1)
	mr := mixed[0].GetReadings()
	require.NotEmpty(t, mr, "mixed-case should receive lower tags (no isMixedCase guard)")
	require.Equal(t, "BaRa", mr[0].GetToken())

	// Unknown → Tag() null POS.
	unk := tagger.Tag([]string{"xyzzyqqq"})
	require.Len(t, unk, 1)
	require.Len(t, unk[0].GetReadings(), 1)
	require.Nil(t, unk[0].GetReadings()[0].GetPOSTag())

	// UTF-16 pos: two tokens "eo" (2) + "ti" (2)
	posOut := tagger.Tag([]string{"eo", "ti"})
	require.Len(t, posOut, 2)
	require.Equal(t, 0, posOut[0].GetStartPos())
	require.Equal(t, 2, posOut[1].GetStartPos())
}

// myAssertTagger ports Java TestTools.myAssert(input, expected, tokenizer, tagger):
// tokenize → drop non-word tokens → tag → sorted readings joined by " -- ".
func myAssertTagger(input string) string {
	EnsureDefaultBretonTagger()
	tagger := DefaultBretonTagger
	// Java Breton language uses BretonWordTokenizer.
	wt := brtok.NewBretonWordTokenizer()
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
