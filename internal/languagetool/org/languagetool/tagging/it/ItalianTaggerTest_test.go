package it

// Twin of languagetool-language-modules/it/src/test/java/org/languagetool/tagging/it/ItalianTaggerTest.java
import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of ItalianTaggerTest.testTagger
// (Java TestTools.myAssert with WordTokenizer + ItalianTagger).
func TestItalianTagger_Tagger(t *testing.T) {
	if DiscoverItalianPOSDict() == "" {
		t.Skip("italian.dict not in tree")
	}
	EnsureDefaultItalianTagger()
	require.NotNil(t, DefaultItalianTagger)
	require.NotNil(t, DefaultItalianTagger.GetWordTagger())

	// Java ItalianTaggerTest.testTagger expected strings (readings sorted in TestTools).
	cases := []struct {
		input string
		want  string
	}{
		{
			"Non c'è linguaggio senza inganno.",
			"Non/[non]ADV -- c/[C]NPR -- è/[essere]AUX:ind+pres+3+s|è/[essere]VER:ind+pres+3+s -- linguaggio/[linguaggio]NOUN-M:s -- senza/[senza]CON|senza/[senza]PRE -- inganno/[ingannare]VER:ind+pres+1+s|inganno/[inganno]NOUN-M:s",
		},
		{
			"Amo quelli che desiderano l'impossibile.",
			"Amo/[amare]VER:ind+pres+1+s -- quelli/[quelli]PRO-DEMO-M-P|quelli/[quello]DET-DEMO:m+p -- che/[che]CON|che/[che]DET-WH:f+p|che/[che]DET-WH:f+s|che/[che]DET-WH:m+p|che/[che]DET-WH:m+s|che/[che]WH-CHE -- desiderano/[desiderare]VER:ind+pres+3+p -- l/[null]null -- impossibile/[impossibile]ADJ:pos+f+s|impossibile/[impossibile]ADJ:pos+m+s",
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

// Twin of ItalianTaggerTest.testDictionary (Java TestTools.testDictionary).
// Java walks every Morfologik WordData and only warns on empty POS; no assert fail.
// Full FSA DictionaryLookup iteration is not yet ported; this opens the real
// italian.dict and checks sample surfaces all carry non-empty POS tags.
func TestItalianTagger_Dictionary(t *testing.T) {
	if DiscoverItalianPOSDict() == "" {
		t.Skip("italian.dict not in tree")
	}
	EnsureDefaultItalianTagger()
	tagger := DefaultItalianTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, ItalianDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, ItalianPOSDictPath(), "real italian.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	// Sample of surfaces present in Java testTagger / lexicon — each must have POS.
	// "c" is lowercase; dict stores "C"/NPR — BaseTagger TagWordExact is exact only.
	// TagWordExact used for dict proof of binary entries.
	samples := []string{"non", "C", "è", "linguaggio", "senza", "inganno", "amo",
		"quelli", "che", "desiderano", "impossibile", "essere", "amare", "desiderare"}
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
	EnsureDefaultItalianTagger()
	tagger := DefaultItalianTagger
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
