package da

// Outcome-behavior matrix for org.languagetool.tagging.da.DanishTagger.
// Upstream has no DanishTaggerTest.java; expectations derived from the same
// official danish.dict + BaseTagger case-merge (TestTools.myAssert style).

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// TestDanishTagger_DictionaryPath asserts Java DanishTagger super path constant.
func TestDanishTagger_DictionaryPath(t *testing.T) {
	tagger := NewDanishTagger(nil)
	require.Equal(t, "/da/danish.dict", DanishDictPath)
	require.Equal(t, DanishDictPath, tagger.GetDictionaryPath())
	require.Equal(t, "da", tagger.LocaleLanguage)
	require.True(t, tagger.TagLowercaseWithUppercase,
		"Java BaseTagger(filename, Locale) defaults tagLowercaseWithUppercase=true")
}

// TestDanishTagger_Tagger asserts myAssert-style outcomes for Danish surfaces
// present in official danish.dict (readings sorted like TestTools.myAssert).
func TestDanishTagger_Tagger(t *testing.T) {
	if DiscoverDanishPOSDict() == "" {
		t.Skip("danish.dict not in tree")
	}
	EnsureDefaultDanishTagger()
	require.NotNil(t, DefaultDanishTagger)
	require.NotNil(t, DefaultDanishTagger.GetWordTagger())
	require.NotEmpty(t, DanishPOSDictPath(), "real danish.dict must load")

	cases := []struct {
		input string
		want  string
	}{
		{
			"Det er en god dag",
			"Det/[et]art -- er/[være]ver:præ:akt -- en/[en]art -- god/[god]adj:ube:sin:utr:pos -- dag/[dag]sub:ube:sin:utr:nom",
		},
		{
			"slidbane er stor",
			"slidbane/[slidbane]sub:ube:sin:utr:nom -- er/[være]ver:præ:akt -- stor/[stor]adj:ube:sin:utr:pos",
		},
		{
			"Huset er lille",
			"Huset/[hus]sub:bes:sin:neu:nom|Huset/[huse]ver:kor:akt -- er/[være]ver:præ:akt -- lille/[lille]adj:bes:sin:neu:pos|lille/[lille]adj:bes:sin:utr:pos|lille/[lille]adj:ube:sin:neu:pos|lille/[lille]adj:ube:sin:utr:pos",
		},
		{
			"hun var her",
			"hun/[hun]pron:sin:nom|hun/[hun]sub:ube:sin:utr:nom -- var/[vare]ver:imp:akt|var/[være]ver:dat:akt -- her/[her]adv",
		},
		{
			"vi er der",
			"vi/[vi]pron:plu:nom|vi/[vi]sub:ube:sin:neu:nom|vi/[vi]ver:imp:akt|vi/[vi]ver:inf:akt -- er/[være]ver:præ:akt -- der/[der]adv",
		},
		// Unknown surface → null POS reading (BaseTagger empty → invent null token).
		{
			"xyzzyqqq",
			"xyzzyqqq/[null]null",
		},
		// "har" is common in Danish but absent from POS dict → null POS.
		{
			"Jeg har ikke tid",
			"Jeg/[jeg]pron:sin:nom|Jeg/[jeg]sub:ube:sin:neu:nom -- har/[null]null -- ikke/[ikke]adv -- tid/[tid]sub:ube:plu:utr:nom|tid/[tid]sub:ube:sin:utr:nom",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// TestDanishTagger_Dictionary opens the real danish.dict and checks sample
// surfaces all carry non-empty POS tags (TestTools.testDictionary spirit).
func TestDanishTagger_Dictionary(t *testing.T) {
	if DiscoverDanishPOSDict() == "" {
		t.Skip("danish.dict not in tree")
	}
	EnsureDefaultDanishTagger()
	tagger := DefaultDanishTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, DanishDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, DanishPOSDictPath(), "real danish.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	// Sample surfaces present in danish.dict (TagWordExact = exact FSA key).
	samples := []string{
		"er", "jeg", "i", "en", "at", "det", "ikke", "du", "han", "til", "på",
		"for", "og", "et", "af", "med", "mig", "den", "hun", "var", "kan", "der",
		"vi", "de", "om", "hvad", "så", "som", "hvor", "ved", "meget", "her",
		"ham", "sig", "være", "fra", "dag", "slidbane", "huset", "hus", "mand",
		"kvinde", "stor", "lille", "godt", "tid", "god",
	}
	for _, w := range samples {
		tw := tagger.TagWordExact(w)
		require.NotEmpty(t, tw, "dict entry missing for %q", w)
		for _, tword := range tw {
			require.NotEmpty(t, tword.PosTag, "**** Warning-equivalent: %s/%s lacks a POS tag", w, tword.Lemma)
		}
	}
}

// TestDanishTagger_CaseMerge sentence-start capital gets lower tags when not mixed case.
func TestDanishTagger_CaseMerge(t *testing.T) {
	if DiscoverDanishPOSDict() == "" {
		t.Skip("danish.dict not in tree")
	}
	EnsureDefaultDanishTagger()
	tagger := DefaultDanishTagger

	// "Det" is not stored uppercase; BaseTagger merges lower "det" → et|art.
	tw := tagger.TagWord("Det")
	require.NotEmpty(t, tw)
	found := false
	for _, r := range tw {
		if r.Lemma == "et" && r.PosTag == "art" {
			found = true
		}
	}
	require.True(t, found, "sentence-start Det should pick up det/et|art lower tags: %#v", tw)

	// Mixed case must not merge lower tags (BaseTagger isMixedCase guard).
	// "HuSet" is mixed → only exact surface lookup (empty) → no lower merge.
	mixed := tagger.TagWord("HuSet")
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
	EnsureDefaultDanishTagger()
	tagger := DefaultDanishTagger
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
