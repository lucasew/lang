package sr

// Twin of AbstractSerbianTaggerTest (Java has no @Test methods).
// Provides assertHasLemmaAndPos + myAssert helpers for Ekavian/Jekavian tests.

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// assertHasLemmaAndPos ports AbstractSerbianTaggerTest.assertHasLemmaAndPos.
// lemma or posTag empty string means "don't care" (Java null).
func assertHasLemmaAndPos(t *testing.T, tagger *SerbianTagger, inflected, lemma, posTag string) {
	t.Helper()
	tags := tagger.Tag([]string{inflected})
	var allTags strings.Builder
	found := false
	for _, atr := range tags {
		if atr == nil {
			continue
		}
		for _, token := range atr.GetReadings() {
			if token == nil {
				continue
			}
			crtLemma, crtPOS := "", ""
			if token.GetLemma() != nil {
				crtLemma = *token.GetLemma()
			}
			if token.GetPOSTag() != nil {
				crtPOS = *token.GetPOSTag()
			}
			allTags.WriteString(fmt.Sprintf("[%s/%s]", crtLemma, crtPOS))
			lemmaOK := lemma == "" || lemma == crtLemma
			posOK := posTag == "" || posTag == crtPOS
			if lemmaOK && posOK {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	require.True(t, found, "Lemma and POS not found for word [%s]! Expected [%s/%s]. Actual: %s",
		inflected, lemma, posTag, allTags.String())
}

// myAssertTagger ports Java TestTools.myAssert(input, expected, tokenizer, tagger):
// tokenize → drop non-word tokens → tag → sorted readings joined by " -- ".
func myAssertTagger(tagger *SerbianTagger, input string) string {
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
