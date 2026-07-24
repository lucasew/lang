package tl

// Outcome-behavior matrix for org.languagetool.tagging.tl.TagalogTagger.
// Upstream has no TagalogTaggerTest.java; expectations derived from the same
// official tagalog.dict + BaseTagger case-merge (TestTools.myAssert style).

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// TestTagalogTagger_DictionaryPath asserts Java TagalogTagger super path constant.
func TestTagalogTagger_DictionaryPath(t *testing.T) {
	tagger := NewTagalogTagger(nil)
	require.Equal(t, "/tl/tagalog.dict", TagalogDictPath)
	require.Equal(t, TagalogDictPath, tagger.GetDictionaryPath())
	require.Equal(t, "tl", tagger.LocaleLanguage)
	require.True(t, tagger.TagLowercaseWithUppercase,
		"Java BaseTagger(filename, Locale) defaults tagLowercaseWithUppercase=true")
}

// TestTagalogTagger_Tagger asserts myAssert-style outcomes for Tagalog surfaces
// present in official tagalog.dict (readings sorted like TestTools.myAssert).
func TestTagalogTagger_Tagger(t *testing.T) {
	if DiscoverTagalogPOSDict() == "" {
		t.Skip("tagalog.dict not in tree")
	}
	EnsureDefaultTagalogTagger()
	require.NotNil(t, DefaultTagalogTagger)
	require.NotNil(t, DefaultTagalogTagger.GetWordTagger())
	require.NotEmpty(t, TagalogPOSDictPath(), "real tagalog.dict must load")

	cases := []struct {
		input string
		want  string
	}{
		{
			"ako ay tao",
			"ako/[ako]PANP ST S -- ay/[ay]INTR|ay/[ay]MALM -- tao/[tao]NCOM 1",
		},
		{
			"siya ay bata",
			"siya/[siya]PANP RD S -- ay/[ay]INTR|ay/[ay]MALM -- bata/[bata]ADUN|bata/[bata]NCOM 1",
		},
		{
			"ang bahay",
			"ang/[ang]DECN NOM|ang/[ang]NA -- bahay/[bahay]NCOM 2",
		},
		{
			"kumain ako",
			"kumain/[kumain]VACF CM B -- ako/[ako]PANP ST S",
		},
		{
			"hindi ko alam",
			"hindi/[hindi]AVGI AL|hindi/[hindi]IRID|hindi/[hindi]MANE -- ko/[ko]PNGP ST S -- alam/[alam]VOTF OT B",
		},
		{
			"sila ay magaling",
			"sila/[sila]PANP RD P -- ay/[ay]INTR|ay/[ay]MALM -- magaling/[magaling]ADMO S|magaling/[magaling]AVMA AL",
		},
		{
			"sa mga araw",
			"sa/[sa]DECN DAT|sa/[sa]NA|sa/[sa]NCOM 2 -- mga/[mga]DEPL NUL|mga/[mga]NA -- araw/[araw]NCOM 2",
		},
		{
			"siya ay lalaki",
			"siya/[siya]PANP RD S -- ay/[ay]INTR|ay/[ay]MALM -- lalaki/[lalaki]NCOM 1",
		},
		{
			"ang babae",
			"ang/[ang]DECN NOM|ang/[ang]NA -- babae/[babae]NCOM 1",
		},
		{
			"pumunta sila",
			"pumunta/[pumunta]VACF CM B -- sila/[sila]PANP RD P",
		},
		// Unknown surface → null POS reading (BaseTagger empty → invent null token).
		{
			"xyzzyqqq",
			"xyzzyqqq/[null]null",
		},
		// Sentence-start capital merges lower tags (tagLowercaseWithUppercase=true).
		{
			"Ako",
			"Ako/[ako]PANP ST S",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// TestTagalogTagger_Dictionary opens the real tagalog.dict and checks sample
// surfaces all carry non-empty POS tags (TestTools.testDictionary spirit).
func TestTagalogTagger_Dictionary(t *testing.T) {
	if DiscoverTagalogPOSDict() == "" {
		t.Skip("tagalog.dict not in tree")
	}
	EnsureDefaultTagalogTagger()
	tagger := DefaultTagalogTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, TagalogDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, TagalogPOSDictPath(), "real tagalog.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	// Sample surfaces present in tagalog.dict (TagWordExact = exact FSA key).
	samples := []string{
		"sa", "ng", "na", "ko", "mga", "ay", "ako", "si", "siya", "hindi",
		"ka", "niya", "nang", "ni", "ba", "isang", "ito", "kung", "araw",
		"bahay", "tao", "bata", "babae", "lalaki", "kumain", "pumunta",
		"kailangan", "lahat", "sila", "kami", "tayo", "wala", "alam",
		"magaling", "oras", "bakit", "saan", "ano", "isa", "ang", "Tagalog",
	}
	for _, w := range samples {
		tw := tagger.TagWordExact(w)
		require.NotEmpty(t, tw, "dict entry missing for %q", w)
		for _, tword := range tw {
			require.NotEmpty(t, tword.PosTag, "**** Warning-equivalent: %s/%s lacks a POS tag", w, tword.Lemma)
		}
	}
}

// TestTagalogTagger_CaseMerge sentence-start capital gets lower tags when not mixed case.
func TestTagalogTagger_CaseMerge(t *testing.T) {
	if DiscoverTagalogPOSDict() == "" {
		t.Skip("tagalog.dict not in tree")
	}
	EnsureDefaultTagalogTagger()
	tagger := DefaultTagalogTagger

	// "Ako" is not stored uppercase; BaseTagger merges lower "ako" → ako|PANP ST S.
	tw := tagger.TagWord("Ako")
	require.NotEmpty(t, tw)
	found := false
	for _, r := range tw {
		if r.Lemma == "ako" && r.PosTag == "PANP ST S" {
			found = true
		}
	}
	require.True(t, found, "sentence-start Ako should pick up ako/PANP ST S lower tags: %#v", tw)

	// Mixed case must not merge lower tags (BaseTagger isMixedCase guard).
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
	EnsureDefaultTagalogTagger()
	tagger := DefaultTagalogTagger
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
