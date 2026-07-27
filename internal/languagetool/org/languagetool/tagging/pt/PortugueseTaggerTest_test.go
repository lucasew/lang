package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/tagging/pt/PortugueseTaggerTest.java
//
// Official portuguese.dict + PortugueseWordTokenizer + PortugueseTagger.tag
// (Java TestTools.myAssert expected strings).

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	pttok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/pt"
	"github.com/stretchr/testify/require"
)

func requirePortugueseDict(t *testing.T) {
	t.Helper()
	// Official portuguese.dict under third_party/portuguese-pos-dict (Java /pt/portuguese.dict).
	// No invent Skip — same bar as PolishWordTokenizer close-out with real dict.
	require.NotEmpty(t, DiscoverPortuguesePOSDict(),
		"official portuguese.dict must be discoverable (third_party/portuguese-pos-dict or LANG_PORTUGUESE_DICT)")
	EnsureDefaultPortugueseTagger()
	require.NotNil(t, DefaultPortugueseTagger, "PortugueseTagger must load from official portuguese.dict")
	require.NotNil(t, DefaultPortugueseTagger.GetWordTagger())
	require.NotEmpty(t, PortuguesePOSDictPath(), "real portuguese.dict must load")
}

// Twin of PortugueseTaggerTest.testDictionary (Java TestTools.testDictionary).
// Java walks every Morfologik WordData and only warns on empty POS; no assert fail.
// Full FSA DictionaryLookup iteration is not yet ported; this opens the real
// portuguese.dict and checks sample surfaces all carry non-empty POS tags.
func TestPortugueseTagger_Dictionary(t *testing.T) {
	requirePortugueseDict(t)
	tagger := DefaultPortugueseTagger
	require.Equal(t, PortugueseDictPath, tagger.GetDictionaryPath())
	require.NotNil(t, tagger.GetWordTagger())

	// Sample of surfaces present in Java testTagger / lexicon — each must have POS.
	samples := []string{
		"estes", "são", "os", "meus", "amigos", "tentou", "resolver",
		"deixe-me", "das", "to", "ao", "jiu-jitsu", "trepei",
	}
	for _, w := range samples {
		tw := tagger.TagWordExact(w)
		require.NotEmpty(t, tw, "dict entry missing for %q", w)
		for _, tword := range tw {
			require.NotEmpty(t, tword.PosTag, "**** Warning-equivalent: %s/%s lacks a POS tag", w, tword.Lemma)
		}
	}
}

// Twin of path/ctor: Java PortugueseTagger uses /pt/portuguese.dict.
func TestPortugueseTagger_DictionaryPath(t *testing.T) {
	tagger := NewPortugueseTagger(nil)
	require.Equal(t, PortugueseDictPath, tagger.GetDictionaryPath())
	require.False(t, tagger.OverwriteWithManualTagger())
}

// Twin of PortugueseTaggerTest.testTagger
func TestPortugueseTagger_Tagger(t *testing.T) {
	requirePortugueseDict(t)
	cases := []struct {
		input string
		want  string
	}{
		{
			"Estes são os meus amigos.",
			"Estes/[este]AQ0CP0|Estes/[este]DD0MP0|Estes/[este]NCMP000|Estes/[este]PD0MP000 -- " +
				"são/[ser]VMIP3P0|são/[são]AQ0MS0|são/[são]NCMS000 -- " +
				"os/[ele]PP3MPA00|os/[o]DA0MP0|os/[o]PD0MP000 -- " +
				"meus/[meu]AP0MP1S|meus/[meu]DP1MPS -- " +
				"amigos/[amigo]AQ0MP0|amigos/[amigo]NCMP000",
		},
		{
			"tentou resolver",
			"tentou/[tentar]VMIS3S0 -- resolver/[resolver]VMN0000|resolver/[resolver]VMN01S0|resolver/[resolver]VMN03S0|resolver/[resolver]VMSF1S0|resolver/[resolver]VMSF3S0",
		},
		// as of dict v0.13, verbs with enclitics are considered as a single word and tagged as such
		{
			"Deixe-me",
			"Deixe-me/[deixar]VMM03S0:PP1CSO00|Deixe-me/[deixar]VMSP1S0:PP1CSO00|Deixe-me/[deixar]VMSP3S0:PP1CSO00",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// Twin of PortugueseTaggerTest.testTaggerTagsOrdinalAbbreviations
func TestPortugueseTagger_TaggerTagsOrdinalAbbreviations(t *testing.T) {
	requirePortugueseDict(t)
	cases := []struct {
		input string
		want  string
	}{
		// with regular lowercase letters
		{"1.as", "1.as/[1.º]AO0FP0|1.as/[1.º]NCFP000"},
		{"2as", "2as/[2º]AO0FP0|2as/[2º]NCFP000"},
		{"300as", "300as/[300º]AO0FP0|300as/[300º]NCFP000"},
		// with ordinal indicator
		{"4.ªˢ", "4.ªˢ/[4.º]AO0FP0|4.ªˢ/[4.º]NCFP000"},
		{"5ªˢ", "5ªˢ/[5º]AO0FP0|5ªˢ/[5º]NCFP000"},
		{"600ªˢ", "600ªˢ/[600º]AO0FP0|600ªˢ/[600º]NCFP000"},
		// with superscript
		{"7.ᵃˢ", "7.ᵃˢ/[7.º]AO0FP0|7.ᵃˢ/[7.º]NCFP000"},
		{"8ᵃˢ", "8ᵃˢ/[8º]AO0FP0|8ᵃˢ/[8º]NCFP000"},
		{"900ᵃˢ", "900ᵃˢ/[900º]AO0FP0|900ᵃˢ/[900º]NCFP000"},
		// percent sign
		{"10%", "10%/[10%]NCMP000"},
		{"−11.000%", "−11.000%/[−11.000%]NCMP000"},
		// degree sign
		{"12°", "12°/[12°]NCMP000"},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// Twin of PortugueseTaggerTest.testContractionTagging
func TestPortugueseTagger_ContractionTagging(t *testing.T) {
	requirePortugueseDict(t)
	cases := []struct {
		input string
		want  string
	}{
		{"das", "das/[de:o]SPS00:DA0FP0"},
		{"to", "to/[tu:ele]PP2CSO00:PP3MSA00"},
		{"ao", "ao/[a:o]SPS00:DA0MS0"},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// Twin of PortugueseTaggerTest.testTaggerTagsCompoundsRegardlessOfLetterCase
func TestPortugueseTagger_TaggerTagsCompoundsRegardlessOfLetterCase(t *testing.T) {
	requirePortugueseDict(t)
	cases := []struct {
		input string
		want  string
	}{
		{"jiu-jitsu", "jiu-jitsu/[jiu-jitsu]NCMS000"},
		{"Jiu-jitsu", "Jiu-jitsu/[jiu-jitsu]NCMS000"},
		{"JIU-JITSU", "JIU-JITSU/[jiu-jitsu]NCMS000"},
		{"Jiu-Jitsu", "Jiu-Jitsu/[jiu-jitsu]NCMS000"},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// Twin of PortugueseTaggerTest.testTagProductivePrefixesNotPresentInSpeller
func TestPortugueseTagger_TagProductivePrefixesNotPresentInSpeller(t *testing.T) {
	requirePortugueseDict(t)
	cases := []struct {
		input string
		want  string
	}{
		// not a real prefix, must be null
		{"xoxotrepei", "xoxotrepei/[null]null"},
		// new prefixes, to work with dict v0.13 (soto- requires hyphen)
		{"soto-trepei", "soto-trepei/[soto-trepar]VMIS1S0"},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// Twin of PortugueseTaggerTest.testTaggerTagsVerbsWithEnclitics
func TestPortugueseTagger_TaggerTagsVerbsWithEnclitics(t *testing.T) {
	requirePortugueseDict(t)
	// requires dictionary v0.14
	got := myAssertTagger("distribuí-lo-ei")
	require.Equal(t, "distribuí-lo-ei/[distribuir]VMIF1S0:PP3MSA00", got)
}

// myAssertTagger ports Java TestTools.myAssert(input, expected, PortugueseWordTokenizer, tagger):
// tokenize → drop non-word tokens → tag → sorted readings joined by " -- ".
func myAssertTagger(input string) string {
	EnsureDefaultPortugueseTagger()
	tagger := DefaultPortugueseTagger
	wt := pttok.NewPortugueseWordTokenizer()
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
