package eo

// Twin of languagetool-language-modules/eo/src/test/java/org/languagetool/tagging/eo/EsperantoTaggerTest.java

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of EsperantoTaggerTest.testTagger
// (Java TestTools.myAssert with WordTokenizer + EsperantoTagger).
func TestEsperantoTagger_Tagger(t *testing.T) {
	if !EOResourcesAvailable() {
		t.Fatal("official EO resources not found (manual-tagger.txt, verb-tr.txt, verb-ntr.txt, root-ant-at.txt)")
	}
	// Java EsperantoTaggerTest.testTagger expected strings (readings sorted in TestTools).
	cases := []struct {
		input string
		want  string
	}{
		{
			"Tio estas simpla testo",
			"Tio/[null]T nak np t o -- estas/[esti]V nt as -- simpla/[simpla]A nak np -- testo/[testo]O nak np",
		},
		{
			"Mi malsategas",
			"Mi/[mi]R nak np -- malsategas/[malsategi]V nt as",
		},
		{
			"Li malŝategas sin",
			"Li/[li]R nak np -- malŝategas/[malŝategi]V tr as -- sin/[si]R akz np",
		},
		// Esperanto pangram: all letters; lemma transformed from x-system into Unicode.
		{
			"Sxajnas ke sagaca monahxo lauxtvocxe rifuzadis pregxi sur herbajxo",
			"Sxajnas/[ŝajni]V nt as -- " +
				"ke/[ke]_ -- " +
				"sagaca/[sagaca]A nak np -- " +
				"monahxo/[monaĥo]O nak np -- " +
				"lauxtvocxe/[laŭtvoĉe]E nak -- " +
				"rifuzadis/[rifuzadi]V tr is -- " +
				"pregxi/[preĝi]V nt i -- " +
				"sur/[sur]P kak -- " +
				"herbajxo/[herbaĵo]O nak np",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// myAssertTagger ports Java TestTools.myAssert(input, expected, tokenizer, tagger):
// tokenize → drop non-word tokens → tag → sorted readings joined by " -- ".
func myAssertTagger(input string) string {
	tagger := NewEsperantoTagger()
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
