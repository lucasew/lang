package ga

// Twin of languagetool-language-modules/ga/src/test/java/org/languagetool/tagging/ga/IrishTaggerTest.java
//
// Official irish.dict + WordTokenizer + IrishTagger.tag
// (Java TestTools.myAssert expected strings).

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

func requireIrishDict(t *testing.T) {
	t.Helper()
	if DiscoverIrishPOSDict() == "" {
		t.Skip("irish.dict not found (LANG_IRISH_DICT / LANG_GA_DICT / /tmp/ga-dict-fetch / third_party)")
	}
	EnsureDefaultIrishTagger()
	require.NotNil(t, DefaultIrishTagger)
	require.NotNil(t, DefaultIrishTagger.GetWordTagger())
	require.NotEmpty(t, IrishPOSDictPath(), "real irish.dict must load")
}

// Twin of path/ctor: Java IrishTagger uses /ga/irish.dict.
func TestIrishTagger_DictionaryPath(t *testing.T) {
	tagger := NewIrishTagger(nil)
	require.Equal(t, IrishDictPath, tagger.GetDictionaryPath())
	require.False(t, tagger.OverwriteWithManualTagger())
}

// Twin of IrishTaggerTest.testDictionary (Java TestTools.testDictionary).
// Java walks every Morfologik WordData and only warns on empty POS; no assert fail.
// Full FSA DictionaryLookup iteration is not yet ported; this opens the real
// irish.dict and checks sample surfaces all carry non-empty POS tags.
func TestIrishTagger_Dictionary(t *testing.T) {
	requireIrishDict(t)
	tagger := DefaultIrishTagger
	require.Equal(t, IrishDictPath, tagger.GetDictionaryPath())
	require.NotNil(t, tagger.GetWordTagger())

	// Sample of surfaces present in Java testTagger / lexicon — each must have POS.
	samples := []string{
		"bhí", "sé", "anseo", "athair", "t-athair", "seomra",
		"bí", "is",
	}
	for _, w := range samples {
		tw := tagger.TagWordExact(w)
		require.NotEmpty(t, tw, "dict entry missing for %q", w)
		for _, tword := range tw {
			require.NotEmpty(t, tword.PosTag, "**** Warning-equivalent: %s/%s lacks a POS tag", w, tword.Lemma)
		}
	}
}

// Twin of IrishTaggerTest.testTagger (Java TestTools.myAssert cases).
func TestIrishTagger_Tagger(t *testing.T) {
	requireIrishDict(t)
	cases := []struct {
		input string
		want  string
	}{
		{
			"nonexistingword",
			"nonexistingword/[null]null",
		},
		{
			"bhí sé anseo",
			"bhí/[bí]Noun:Fem:Com:Sg:DefArt|bhí/[bí]Noun:Fem:Com:Sg:Len|bhí/[bí]Noun:Fem:Gen:Sg:Len|bhí/[bí]Noun:Fem:Voc:Sg:Len|bhí/[bí]Verb:VI:PastInd:Len -- " +
				"sé/[is]Cop:Pres:Pron:Pers:3P:Sg:Masc|sé/[sé]Noun:Masc:Com:Sg|sé/[sé]Noun:Masc:Com:Sg:DefArt|sé/[sé]Noun:Masc:Com:Sg:Ecl|sé/[sé]Noun:Masc:Gen:Sg|sé/[sé]Noun:Masc:Gen:Sg:Ecl|sé/[sé]Num:Card|sé/[sé]Num:Card:Ecl|sé/[sé]Pron:Pers:3P:Sg:Masc:Sbj -- " +
				"anseo/[anseo]Adv:Loc",
		},
		{
			"t-athair",
			"t-athair/[athair]Noun:Masc:Com:Sg:DefArt",
		},
		{
			"tAthair",
			"tAthair/[athair]Noun:Masc:Com:Sg:DefArt",
		},
		{
			"nAthair",
			"nAthair/[athair]Noun:Fem:Com:Sg:Ecl|nAthair/[athair]Noun:Masc:Com:Sg:Ecl",
		},
		{
			"t-seomra",
			"t-seomra/[null]null|t-seomra/[seomra]Noun:Masc:Com:Sg:DefArt:MorphError",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// myAssertTagger ports Java TestTools.myAssert(input, expected, WordTokenizer, tagger):
// tokenize → drop non-word tokens → tag → sorted readings joined by " -- ".
func myAssertTagger(input string) string {
	EnsureDefaultIrishTagger()
	tagger := DefaultIrishTagger
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
