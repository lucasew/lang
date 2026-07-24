package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/tagging/pl/PolishTaggerTest.java
import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of PolishTaggerTest.testTagger
// (Java TestTools.myAssert with WordTokenizer + PolishTagger).
func TestPolishTagger_Tagger(t *testing.T) {
	if DiscoverPolishPOSDict() == "" {
		t.Skip("polish.dict not in tree")
	}
	EnsureDefaultPolishTagger()
	require.NotNil(t, DefaultPolishTagger)
	require.NotNil(t, DefaultPolishTagger.GetWordTagger())

	// Java PolishTaggerTest.testTagger expected strings (readings sorted in TestTools).
	cases := []struct {
		input string
		want  string
	}{
		{
			"To jest duży dom.",
			"To/[ten]adj:sg:acc:n1.n2:pos|To/[ten]adj:sg:nom.voc:n1.n2:pos|To/[to]conj|To/[to]qub|To/[to]subst:sg:acc:n2|To/[to]subst:sg:nom:n2 -- jest/[być]verb:fin:sg:ter:imperf:nonrefl -- duży/[duży]adj:sg:acc:m3:pos|duży/[duży]adj:sg:nom.voc:m1.m2.m3:pos -- dom/[dom]subst:sg:acc:m3|dom/[dom]subst:sg:nom:m3",
		},
		{
			"Krowa pasie się na pastwisku.",
			"Krowa/[krowa]subst:sg:nom:f -- pasie/[pas]subst:sg:loc:m3|pasie/[pas]subst:sg:voc:m3|pasie/[paść]verb:fin:sg:ter:imperf:refl.nonrefl -- się/[się]qub|się/[się]siebie:acc:nakc|się/[się]siebie:gen:nakc -- na/[na]interj|na/[na]prep:acc|na/[na]prep:loc -- pastwisku/[pastwisko]subst:sg:dat:n2|pastwisku/[pastwisko]subst:sg:loc:n2",
		},
		{
			"blablabla",
			"blablabla/[null]null",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// Twin of PolishTaggerTest.testDictionary (Java TestTools.testDictionary).
// Java walks every Morfologik WordData and only warns on empty POS; no assert fail.
// Full FSA DictionaryLookup iteration is not yet ported; this opens the real
// polish.dict and checks sample surfaces all carry non-empty POS tags.
func TestPolishTagger_Dictionary(t *testing.T) {
	if DiscoverPolishPOSDict() == "" {
		t.Skip("polish.dict not in tree")
	}
	EnsureDefaultPolishTagger()
	tagger := DefaultPolishTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, PolishDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, PolishPOSDictPath(), "real polish.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	// Sample of surfaces present in Java testTagger / lexicon — each must have POS.
	samples := []string{"to", "jest", "duży", "dom", "krowa", "pasie", "się", "na", "pastwisku",
		"ten", "być", "pas", "paść", "pastwisko"}
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
	EnsureDefaultPolishTagger()
	tagger := DefaultPolishTagger
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
