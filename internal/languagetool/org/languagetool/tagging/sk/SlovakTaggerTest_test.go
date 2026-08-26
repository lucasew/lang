package sk

// Twin of languagetool-language-modules/sk/src/test/java/org/languagetool/tagging/sk/SlovakTaggerTest.java
import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of SlovakTaggerTest.testTagger
// (Java TestTools.myAssert with WordTokenizer + SlovakTagger).
func TestSlovakTagger_Tagger(t *testing.T) {
	if DiscoverSlovakPOSDict() == "" {
		t.Skip("slovak.dict not in tree")
	}
	EnsureDefaultSlovakTagger()
	require.NotNil(t, DefaultSlovakTagger)
	require.NotNil(t, DefaultSlovakTagger.GetWordTagger())

	// Java SlovakTaggerTest.testTagger expected strings (readings sorted in TestTools).
	cases := []struct {
		input string
		want  string
	}{
		{
			"Tu nájdete vybrané čísla a obsahy časopisu Kultúra slova.",
			"Tu/[tu]J|Tu/[tu]PD|Tu/[tu]T -- nájdete/[nájsť]VKdpb+ -- vybrané/[vybraný]Gtfp1x|vybrané/[vybraný]Gtfp4x|vybrané/[vybraný]Gtfp5x|vybrané/[vybraný]Gtip1x|vybrané/[vybraný]Gtip4x|vybrané/[vybraný]Gtip5x|vybrané/[vybraný]Gtnp1x|vybrané/[vybraný]Gtnp4x|vybrané/[vybraný]Gtnp5x|vybrané/[vybraný]Gtns1x|vybrané/[vybraný]Gtns4x|vybrané/[vybraný]Gtns5x -- čísla/[číslo]SSnp1|čísla/[číslo]SSnp4|čísla/[číslo]SSnp5|čísla/[číslo]SSns2 -- a/[a]J|a/[a]O|a/[a]Q|a/[a]SUnp1|a/[a]SUnp2|a/[a]SUnp3|a/[a]SUnp4|a/[a]SUnp5|a/[a]SUnp6|a/[a]SUnp7|a/[a]SUns1|a/[a]SUns2|a/[a]SUns3|a/[a]SUns4|a/[a]SUns5|a/[a]SUns6|a/[a]SUns7|a/[a]T|a/[a]W|a/[as]W -- obsahy/[obsah]SSip1|obsahy/[obsah]SSip4|obsahy/[obsah]SSip5 -- časopisu/[časopis]SSis2|časopisu/[časopis]SSis3 -- Kultúra/[kultúra]SSfs1|Kultúra/[kultúra]SSfs5 -- slova/[slovo]SSns2",
		},
		{
			"blabla",
			"blabla/[null]null",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// Twin of SlovakTaggerTest.testDictionary (Java TestTools.testDictionary).
// Java walks every Morfologik WordData and only warns on empty POS; no assert fail.
// Full FSA DictionaryLookup iteration is not yet ported; this opens the real
// slovak.dict and checks sample surfaces all carry non-empty POS tags.
func TestSlovakTagger_Dictionary(t *testing.T) {
	if DiscoverSlovakPOSDict() == "" {
		t.Skip("slovak.dict not in tree")
	}
	EnsureDefaultSlovakTagger()
	tagger := DefaultSlovakTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, SlovakDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, SlovakPOSDictPath(), "real slovak.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	// Sample of surfaces present in Java testTagger / lexicon — each must have POS.
	// TagWordExact used for dict proof of binary entries (exact surface as stored).
	// Lemmas may differ from surface casing; use forms as stored when possible.
	samples := []string{"tu", "nájdete", "vybrané", "čísla", "a", "obsahy", "časopisu",
		"kultúra", "slova", "nájsť", "vybraný", "číslo", "obsah", "časopis", "slovo"}
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
	EnsureDefaultSlovakTagger()
	tagger := DefaultSlovakTagger
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
