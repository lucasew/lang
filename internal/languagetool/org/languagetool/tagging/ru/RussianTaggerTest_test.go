package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/tagging/ru/RussianTaggerTest.java
import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of RussianTaggerTest.testTagger
// (Java TestTools.myAssert with WordTokenizer + RussianTagger.INSTANCE).
func TestRussianTagger_Tagger(t *testing.T) {
	if DiscoverRussianPOSDict() == "" {
		t.Skip("russian.dict not in tree")
	}
	EnsureDefaultRussianTagger()
	require.NotNil(t, DefaultRussianTagger)
	require.NotNil(t, DefaultRussianTagger.GetWordTagger())

	// Java RussianTaggerTest.testTagger expected strings (readings sorted in TestTools).
	cases := []struct {
		input string
		want  string
	}{
		{
			"Все счастливые семьи похожи друг на друга,  каждая  несчастливая  семья несчастлива по-своему.",
			"Все/[весь]ADJ:MPR:PL:Nom|Все/[весь]ADJ:MPR:PL:V|Все/[все]PNN:PL:Nom|Все/[все]PNN:PL:V|Все/[все]PNN:Sin:Nom|Все/[все]PNN:Sin:V -- счастливые/[счастливый]ADJ:Posit:PL:Nom|счастливые/[счастливый]ADJ:Posit:PL:V -- семьи/[семья]NN:Inanim:Fem:PL:Nom|семьи/[семья]NN:Inanim:Fem:PL:V|семьи/[семья]NN:Inanim:Fem:Sin:R -- похожи/[похожий]ADJ:Short:PL -- друг/[друг]NN:Anim:Masc:Sin:Nom -- на/[на]PREP -- друга/[друг]NN:Anim:Masc:Sin:R|друга/[друг]NN:Anim:Masc:Sin:V -- каждая/[каждый]ADJ:MPR:Fem:Nom -- несчастливая/[несчастливый]ADJ:Posit:Fem:Nom -- семья/[семья]NN:Inanim:Fem:Sin:Nom -- несчастлива/[несчастливый]ADJ:Short:Fem -- по-своему/[по-своему]ADV",
		},
		{
			"Все смешалось в доме Облонских.",
			"Все/[весь]ADJ:MPR:PL:Nom|Все/[весь]ADJ:MPR:PL:V|Все/[все]PNN:PL:Nom|Все/[все]PNN:PL:V|Все/[все]PNN:Sin:Nom|Все/[все]PNN:Sin:V -- смешалось/[смешаться]VB:Past:INTR:PFV:Neut -- в/[в]PREP -- доме/[дом]NN:Inanim:Masc:Sin:P -- Облонских/[null]null",
		},
		{
			"Абдуллаевы",
			"Абдуллаевы/[абдуллаев]NN:Fam:PL:Nom",
		},
		{
			"блукать",
			"блукать/[блукать]VB:INF:",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// Twin of RussianTaggerTest.testDictionary (Java TestTools.testDictionary).
// Java walks every Morfologik WordData and only warns on empty POS; no assert fail.
// Full FSA DictionaryLookup iteration is not yet ported; this opens the real
// russian.dict and checks sample surfaces all carry non-empty POS tags.
func TestRussianTagger_Dictionary(t *testing.T) {
	if DiscoverRussianPOSDict() == "" {
		t.Skip("russian.dict not in tree")
	}
	EnsureDefaultRussianTagger()
	tagger := DefaultRussianTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, RussianDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, RussianPOSDictPath(), "real russian.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	// Sample of surfaces present in Java testTagger / lexicon — each must have POS.
	samples := []string{"все", "счастливые", "семьи", "похожи", "друг", "на", "друга", "каждая",
		"несчастливая", "семья", "несчастлива", "по-своему", "смешалось", "в", "доме",
		"абдуллаевы", "блукать", "дом"}
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
	EnsureDefaultRussianTagger()
	tagger := DefaultRussianTagger
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
