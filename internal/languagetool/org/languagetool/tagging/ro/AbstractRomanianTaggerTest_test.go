package ro

// Twin of AbstractRomanianTaggerTest — shared helpers for RomanianTagger* tests.
// Java: AbstractRomanianTaggerTest.java (createTagger, assertHasLemmaAndPos, testDictionary).

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

// Twin of AbstractRomanianTaggerTest.testDictionary (Java TestTools.testDictionary).
// Full FSA DictionaryLookup iteration is not yet ported; sample surfaces from
// the Java tagger tests must carry non-empty POS tags via the real romanian.dict.
func TestAbstractRomanianTagger_Dictionary(t *testing.T) {
	if DiscoverRomanianPOSDict() == "" {
		t.Skip("romanian.dict not in tree")
	}
	EnsureDefaultRomanianTagger()
	tagger := DefaultRomanianTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, RomanianDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, RomanianPOSDictPath(), "real romanian.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	// Surfaces from RomanianTaggerTest / lexicon — each must have POS.
	// TagWordExact for binary dict proof; manuals covered via TagWord for configurați.
	samplesExact := []string{"mergeam", "merseserăm", "sunt", "este", "frumoasă", "cartea", "fi"}
	for _, w := range samplesExact {
		tw := tagger.TagWordExact(w)
		require.NotEmpty(t, tw, "dict entry missing for %q", w)
		for _, tword := range tw {
			require.NotEmpty(t, tword.PosTag, "**** Warning-equivalent: %s/%s lacks a POS tag", w, tword.Lemma)
		}
	}
	// configurați POS from added.txt (CombiningTagger) — may not be sole binary entry.
	tw := tagger.TagWord("configurați")
	require.NotEmpty(t, tw, "configurați missing (binary+added)")
	found := false
	for _, tword := range tw {
		require.NotEmpty(t, tword.PosTag)
		if tword.Lemma == "configura" && tword.PosTag == "V0p2000cz0" {
			found = true
		}
	}
	require.True(t, found, "configurați must include configura/V0p2000cz0 from added.txt, got %v", tw)
}

// assertHasLemmaAndPos ports AbstractRomanianTaggerTest.assertHasLemmaAndPos.
// lemma or posTag empty string means "don't care" (Java null).
func assertHasLemmaAndPos(t *testing.T, tagger *RomanianTagger, inflected, lemma, posTag string) {
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
func myAssertTagger(tagger *RomanianTagger, input string) string {
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
