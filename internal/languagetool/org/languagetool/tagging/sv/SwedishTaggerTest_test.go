package sv

// Twin of languagetool-language-modules/sv/src/test/java/org/languagetool/tagging/sv/SwedishTaggerTest.java
import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of SwedishTaggerTest.testTagger
// (Java TestTools.myAssert with WordTokenizer + SwedishTagger).
func TestSwedishTagger_Tagger(t *testing.T) {
	if DiscoverSwedishPOSDict() == "" {
		t.Skip("swedish.dict not in tree")
	}
	EnsureDefaultSwedishTagger()
	require.NotNil(t, DefaultSwedishTagger)
	require.NotNil(t, DefaultSwedishTagger.GetWordTagger())

	// Java SwedishTaggerTest.testTagger expected strings (readings sorted in TestTools).
	cases := []struct {
		input string
		want  string
	}{
		{
			"Det är nog bäst att du får en klubba till",
			"Det/[det]PN -- är/[vara]VB:PRS -- nog/[nog]AB -- bäst/[bra]JJ:S|bäst/[bäst]AB|bäst/[god]JJ:S -- att/[att]KN -- du/[du]PN -- får/[få]VB:PRS|får/[får]NN:OF:PLU:NOM:NEU|får/[får]NN:OF:SIN:NOM:NEU -- en/[en]NN:OF:SIN:NOM:UTR|en/[en]PN -- klubba/[klubba]NN:OF:SIN:NOM:UTR|klubba/[klubba]VB:IMP|klubba/[klubba]VB:INF -- till/[till]AB|till/[till]PP",
		},
		{
			// en + passant/[null]null
			"Hon nämnde, en passant, att det inte var klädsamt",
			"Hon/[hon]PN -- nämnde/[nämna]VB:PRT -- en/[en]NN:OF:SIN:NOM:UTR|en/[en]PN -- passant/[null]null -- att/[att]KN -- det/[det]PN -- inte/[inte]AB -- var/[var]AB|var/[var]NN:OF:SIN:NOM:NEU|var/[var]PN|var/[vara]VB:IMP|var/[vara]VB:PRT -- klädsamt/[klädsam]JJ:PN",
		},
		{
			"Nato-vänliga länder har blivit fler.",
			"Nato-vänliga/[null]null -- länder/[land]NN:OF:PLU:NOM:NEU|länder/[länd]NN:OF:PLU:NOM:UTR|länder/[lända]VB:PRS -- har/[ha]VB:PRS -- blivit/[bli]VB:SUP -- fler/[mången]JJ:K",
		},
		{
			"FN:s nya projekt.",
			"FN/[FN]PM:NOM:ACR -- s/[null]null -- nya/[ny]JJ:BF|nya/[ny]JJ:P -- projekt/[projekt]NN:OF:PLU:NOM:NEU|projekt/[projekt]NN:OF:SIN:NOM:NEU",
		},
		{
			"Du menar sannolikt \"massera\" om du inte skriver om masarnas era förstås.",
			"Du/[du]PN -- menar/[mena]VB:PRS -- sannolikt/[sannolik]JJ:PN|sannolikt/[sannolikt]AB -- massera/[massera]VB:IMP|massera/[massera]VB:INF -- om/[om]AB|om/[om]KN|om/[om]PP -- du/[du]PN -- inte/[inte]AB -- skriver/[skriva]VB:PRS -- om/[om]AB|om/[om]KN|om/[om]PP -- masarnas/[mas]NN:BF:PLU:GEN:UTR -- era/[era]NN:OF:SIN:NOM:UTR|era/[era]PN -- förstås/[förstå]VB:INF:PF|förstås/[förstå]VB:PRS:PF|förstås/[förstås]AB",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// Twin of SwedishTaggerTest.testDictionary (Java TestTools.testDictionary).
// Java walks every Morfologik WordData and only warns on empty POS; no assert fail.
// Full FSA DictionaryLookup iteration is not yet ported; this opens the real
// swedish.dict and checks sample surfaces all carry non-empty POS tags.
func TestSwedishTagger_Dictionary(t *testing.T) {
	if DiscoverSwedishPOSDict() == "" {
		t.Skip("swedish.dict not in tree")
	}
	EnsureDefaultSwedishTagger()
	tagger := DefaultSwedishTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, SwedishDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, SwedishPOSDictPath(), "real swedish.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	// Sample of surfaces present in Java testTagger / lexicon — each must have POS.
	// TagWordExact used for dict proof of binary entries (exact surface as stored).
	samples := []string{"det", "är", "vara", "nog", "bäst", "bra", "att", "du", "får", "en",
		"klubba", "till", "hon", "nämnde", "inte", "var", "länder", "har", "blivit", "fler",
		"FN", "nya", "projekt", "menar", "sannolikt", "massera", "skriver", "masarnas", "era", "förstås"}
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
	EnsureDefaultSwedishTagger()
	tagger := DefaultSwedishTagger
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
