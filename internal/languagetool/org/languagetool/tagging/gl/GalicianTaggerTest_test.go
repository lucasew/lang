package gl

// Twin of languagetool-language-modules/gl/src/test/java/org/languagetool/tagging/gl/GalicianTaggerTest.java
import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of GalicianTaggerTest.testTagger
// (Java TestTools.myAssert with WordTokenizer + GalicianTagger).
func TestGalicianTagger_Tagger(t *testing.T) {
	if DiscoverGalicianPOSDict() == "" {
		t.Skip("galician.dict not in tree")
	}
	EnsureDefaultGalicianTagger()
	require.NotNil(t, DefaultGalicianTagger)
	require.NotNil(t, DefaultGalicianTagger.GetWordTagger())
	require.NotEmpty(t, GalicianPOSDictPath(), "real galician.dict must load")

	// Java GalicianTaggerTest.testTagger expected strings (readings sorted in TestTools).
	cases := []struct {
		input string
		want  string
	}{
		{
			"Todo vai mudar",
			"Todo/[todo]DI0MS0|Todo/[todo]PI0MS000 -- vai/[ir]VMIP3S0|vai/[ir]VMM02S0 -- mudar/[mudar]VMN0000|mudar/[mudar]VMN01S0|mudar/[mudar]VMN03S0|mudar/[mudar]VMSF1S0|mudar/[mudar]VMSF3S0",
		},
		{
			"Se aínda somos galegos é por obra e graza do idioma",
			"Se/[se]CS|Se/[se]PP3PN000|Se/[se]PP3SN000 -- aínda/[aínda]CS|aínda/[aínda]RG -- somos/[ser]VSIP1P0 -- galegos/[galego]AQ0MP0|galegos/[galego]NCMP000 -- é/[ser]VSIP3S0 -- por/[por]SPS00 -- obra/[obra]NCFS000|obra/[obrar]VMIP3S0|obra/[obrar]VMM02S0 -- e/[e]CC|e/[e]NCMS000 -- graza/[graza]NCFS000 -- do/[de]SPS00:DA -- idioma/[idioma]NCMS000",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// Twin of path/ctor: Java GalicianTagger uses /gl/galician.dict.
func TestGalicianTagger_DictionaryPath(t *testing.T) {
	tagger := NewGalicianTagger(nil)
	require.Equal(t, GalicianDictPath, tagger.GetDictionaryPath())
	require.False(t, tagger.OverwriteWithManualTagger())
}

// myAssertTagger ports Java TestTools.myAssert(input, expected, tokenizer, tagger):
// tokenize → drop non-word tokens → tag → sorted readings joined by " -- ".
func myAssertTagger(input string) string {
	EnsureDefaultGalicianTagger()
	tagger := DefaultGalicianTagger
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
