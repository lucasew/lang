package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicTaggerTest.java
import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of ArabicTaggerTest.testTagger
// (Java TestTools.myAssert with WordTokenizer + ArabicTagger).
func TestArabicTagger_Tagger(t *testing.T) {
	if DiscoverArabicPOSDict() == "" {
		t.Skip("arabic.dict not in tree")
	}
	EnsureDefaultArabicTagger()
	require.NotNil(t, DefaultArabicTagger)
	require.NotNil(t, DefaultArabicTagger.GetWordTagger())
	require.NotEmpty(t, ArabicPOSDictPath(), "real arabic.dict must load")

	// Java ArabicTaggerTest.testTagger expected strings (readings sorted in TestTools).
	cases := []struct {
		input string
		want  string
	}{
		{
			"الخياريتان",
			"الخياريتان/[خيار]NA-;F3--;--L" +
				"|الخياريتان/[خيار]NJ-;F2--;--L" +
				"|الخياريتان/[خيار]NJ-;F3--;--L",
		},
		{
			"السماء زرقاء",
			"السماء/[سماء]NJ-;F1--;--L|" +
				"السماء/[سماء]NJ-;F1A-;--L|" +
				"السماء/[سماء]NJ-;F1I-;--L|" +
				"السماء/[سماء]NJ-;F1U-;--L" +
				" -- " +
				"زرقاء/[زرقاء]NA-;F1--;---|" +
				"زرقاء/[زرقاء]NA-;F1A-;---|" +
				"زرقاء/[زرقاء]NA-;F1I-;---|" +
				"زرقاء/[زرقاء]NA-;F1U-;---",
		},
		// non-existing-word
		{
			"العباره",
			"العباره/[null]null",
		},
		{
			"والبلاد",
			"والبلاد/[بلاد]NJ-;F3--;W-L|" +
				"والبلاد/[بلاد]NJ-;F3A-;W-L|" +
				"والبلاد/[بلاد]NJ-;F3I-;W-L|" +
				"والبلاد/[بلاد]NJ-;F3U-;W-L|" +
				"والبلاد/[بلاد]NJ-;M1--;W-L|" +
				"والبلاد/[بلاد]NJ-;M1A-;W-L|" +
				"والبلاد/[بلاد]NJ-;M1I-;W-L|" +
				"والبلاد/[بلاد]NJ-;M1U-;W-L",
		},
		{
			"بلادهما",
			"بلادهما/[بلاد]NJ-;F3--;--H|" +
				"بلادهما/[بلاد]NJ-;F3A-;--H|" +
				"بلادهما/[بلاد]NJ-;F3I-;--H|" +
				"بلادهما/[بلاد]NJ-;F3U-;--H|" +
				"بلادهما/[بلاد]NJ-;M1--;--H|" +
				"بلادهما/[بلاد]NJ-;M1A-;--H|" +
				"بلادهما/[بلاد]NJ-;M1I-;--H|" +
				"بلادهما/[بلاد]NJ-;M1U-;--H",
		},
		{
			"وبلادهما",
			"وبلادهما/[بلاد]NJ-;F3--;W-H|" +
				"وبلادهما/[بلاد]NJ-;F3A-;W-H|" +
				"وبلادهما/[بلاد]NJ-;F3I-;W-H|" +
				"وبلادهما/[بلاد]NJ-;F3U-;W-H|" +
				"وبلادهما/[بلاد]NJ-;M1--;W-H|" +
				"وبلادهما/[بلاد]NJ-;M1A-;W-H|" +
				"وبلادهما/[بلاد]NJ-;M1I-;W-H|" +
				"وبلادهما/[بلاد]NJ-;M1U-;W-H",
		},
		{
			"كبلاد",
			"كبلاد/[بلاد]NJ-;F3--;-K-|" +
				"كبلاد/[بلاد]NJ-;F3I-;-K-|" +
				"كبلاد/[بلاد]NJ-;M1--;-K-|" +
				"كبلاد/[بلاد]NJ-;M1I-;-K-",
		},
		{
			"وكالبلاد",
			"وكالبلاد/[بلاد]NJ-;F3--;WKL|" +
				"وكالبلاد/[بلاد]NJ-;F3I-;WKL|" +
				"وكالبلاد/[بلاد]NJ-;M1--;WKL|" +
				"وكالبلاد/[بلاد]NJ-;M1I-;WKL",
		},
		{
			"فاستعملها",
			"فاستعملها/[اِسْتَعْمَلَ]V61;M1H-pa-;W-H|فاستعملها/[اِسْتَعْمَلَ]V61;M1Y-i--;W-H",
		},
		{
			"سيعملون",
			"سيعملون/[أَعْمَلَ]V41;M3H-faU;-S-|" +
				"سيعملون/[أَعْمَلَ]V41;M3H-fpU;-S-|" +
				"سيعملون/[عَمِلَ]V31;M3H-faU;-S-|" +
				"سيعملون/[عَمِلَ]V31;M3H-fpU;-S-|" +
				"سيعملون/[عَمَّلَ]V41;M3H-faU;-S-|" +
				"سيعملون/[عَمَّلَ]V41;M3H-fpU;-S-",
		},
		{
			"فسيعملون",
			"فسيعملون/[أَعْمَلَ]V41;M3H-faU;WS-|" +
				"فسيعملون/[أَعْمَلَ]V41;M3H-fpU;WS-|" +
				"فسيعملون/[عَمِلَ]V31;M3H-faU;WS-|" +
				"فسيعملون/[عَمِلَ]V31;M3H-fpU;WS-|" +
				"فسيعملون/[عَمَّلَ]V41;M3H-faU;WS-|" +
				"فسيعملون/[عَمَّلَ]V41;M3H-fpU;WS-",
		},
		{
			"كتاب",
			"كتاب/[كتاب]NA-;-3--;---|" +
				"كتاب/[كتاب]NA-;-3A-;---|" +
				"كتاب/[كتاب]NA-;-3I-;---|" +
				"كتاب/[كتاب]NA-;-3U-;---|" +
				"كتاب/[كتاب]NJ-;M1--;---|" +
				"كتاب/[كتاب]NJ-;M1A-;---|" +
				"كتاب/[كتاب]NJ-;M1I-;---|" +
				"كتاب/[كتاب]NJ-;M1U-;---|" +
				"كتاب/[كتاب]NM-;M1--;---|" +
				"كتاب/[كتاب]NM-;M1A-;---|" +
				"كتاب/[كتاب]NM-;M1I-;---|" +
				"كتاب/[كتاب]NM-;M1U-;---",
		},
		{
			"للبلاد",
			"للبلاد/[بلاد]NJ-;F3--;-LL|" +
				"للبلاد/[بلاد]NJ-;F3I-;-LL|" +
				"للبلاد/[بلاد]NJ-;M1--;-LL|" +
				"للبلاد/[بلاد]NJ-;M1I-;-LL",
		},
		{
			"للاعب",
			"للاعب/[لاعب]NA-;M1--;-L-|" +
				"للاعب/[لاعب]NA-;M1--;-LL|" +
				"للاعب/[لاعب]NA-;M1I-;-L-|" +
				"للاعب/[لاعب]NA-;M1I-;-LL|" +
				"للاعب/[لَاعَبَ]V41;M1H-pa-;-L-|" +
				"للاعب/[لَاعَبَ]V41;M1Y-i--;-L-",
		},
		{
			"ببلاد",
			"ببلاد/[بلاد]NJ-;F3--;-B-|" +
				"ببلاد/[بلاد]NJ-;F3I-;-B-|" +
				"ببلاد/[بلاد]NJ-;M1--;-B-|" +
				"ببلاد/[بلاد]NJ-;M1I-;-B-",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// Twin of ArabicTaggerTest.testDictionary (Java TestTools.testDictionary).
// Java walks every Morfologik WordData and only warns on empty POS; no assert fail.
// Full FSA DictionaryLookup iteration is not yet ported; this opens the real
// arabic.dict and checks sample surfaces all carry non-empty POS tags.
func TestArabicTagger_Dictionary(t *testing.T) {
	if DiscoverArabicPOSDict() == "" {
		t.Skip("arabic.dict not in tree")
	}
	EnsureDefaultArabicTagger()
	tagger := DefaultArabicTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, ArabicDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, ArabicPOSDictPath(), "real arabic.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	// Sample surfaces from Java testTagger stems / lexicon — each must have POS.
	samples := []string{"كتاب", "بلاد", "سماء", "خيار", "زرقاء", "يعملون", "استعمل", "لاعب", "خياريتان"}
	for _, w := range samples {
		tw := tagger.TagWordExact(w)
		require.NotEmpty(t, tw, "dict entry missing for %q", w)
		for _, tword := range tw {
			require.NotEmpty(t, tword.PosTag, "**** Warning-equivalent: %s/%s lacks a POS tag", w, tword.Lemma)
		}
	}
}

// Twin of path/ctor: Java ArabicTagger uses /ar/arabic.dict.
func TestArabicTagger_DictionaryPath(t *testing.T) {
	tagger := NewArabicTagger(nil)
	require.Equal(t, ArabicDictPath, tagger.GetDictionaryPath())
	require.False(t, tagger.OverwriteWithManualTagger())
}

// myAssertTagger ports Java TestTools.myAssert(input, expected, tokenizer, tagger):
// tokenize → drop non-word tokens → tag → sorted readings joined by " -- ".
func myAssertTagger(input string) string {
	EnsureDefaultArabicTagger()
	tagger := DefaultArabicTagger
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
