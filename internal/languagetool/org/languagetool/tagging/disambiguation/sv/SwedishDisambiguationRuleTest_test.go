package sv

// Twin of SwedishDisambiguationRuleTest.testChunker
// Java: org.languagetool.tagging.disambiguation.sv.SwedishDisambiguationRuleTest#testChunker
// Resources: /sv/multiwords.txt, swedish.dict via SwedishTagger
import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	tagsv "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/sv"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of SwedishDisambiguationRuleTest.testChunker.
// Java: MultiWordChunker.getInstance("/sv/multiwords.txt") → false,false,false defaults;
// WordTokenizer (core), SRXSentenceTokenizer(Swedish), SwedishTagger; TestTools.myAssert.
// VW-skandalen uses tagger-only overload (tokenizer + tagger, " -- " join, no disambiguator).
func TestSwedishDisambiguationRule_Chunker(t *testing.T) {
	if tagsv.DiscoverSwedishPOSDict() == "" {
		t.Skip("swedish.dict not in tree")
	}
	tagsv.EnsureDefaultSwedishTagger()
	require.NotNil(t, tagsv.DefaultSwedishTagger)
	require.NotNil(t, tagsv.DefaultSwedishTagger.GetWordTagger())

	disambiguator := loadSwedishMultiWordChunker(t)

	// Java TestTools.myAssert expected strings (active cases only; commented ones ignored).
	// Readings sorted like TestTools.getAsStrings.
	cases := []struct {
		input string
		want  string
	}{
		{
			"Att testa ... disambiguering",
			"/[null]SENT_START Att/[att]KN  /[null]null testa/[testa]VB:IMP|testa/[testa]VB:INF  /[null]null ./[...]<ELLIPS> ./[null]null ./[...]</ELLIPS>  /[null]null/[null]SENT_START disambiguering/[null]null",
		},
		{
			"Att testa disambiguering är, en passant, kul.",
			"/[null]SENT_START Att/[att]KN  /[null]null testa/[testa]VB:IMP|testa/[testa]VB:INF  /[null]null disambiguering/[null]null  /[null]null är/[vara]VB:PRS ,/[null]null  /[null]null en/[en passant]<NN:OF:SIN:NOM:UTR>|en/[en]NN:OF:SIN:NOM:UTR|en/[en]PN  /[null]null passant/[en passant]</NN:OF:SIN:NOM:UTR> ,/[null]null  /[null]null kul/[kul]JJ:PU ./[null]null",
		},
		{
			"Te från Sri Lanka är mycket gott.",
			"/[null]SENT_START Te/[te]NN:OF:NON:NOM:UTR|Te/[te]NN:OF:SIN:NOM:NEU|Te/[te]VB:IMP|Te/[te]VB:INF  /[null]null från/[från]PP  /[null]null Sri/[Sri Lanka]<PM:NOM>  /[null]null Lanka/[Sri Lanka]</PM:NOM>  /[null]null är/[vara]VB:PRS  /[null]null mycket/[mycken]JJ:PN|mycket/[mycket]AB  /[null]null gott/[god]JJ:PN|gott/[gott]AB ./[null]null",
		},
		{
			"Test ...",
			"/[null]SENT_START Test/[test]NN:OF:PLU:NOM:NEU|Test/[test]NN:OF:SIN:NOM:NEU|Test/[test]NN:OF:SIN:NOM:UTR  /[null]null ./[...]<ELLIPS> ./[null]null ./[...]</ELLIPS>",
		},
		{
			"Test 2 ... ",
			"/[null]SENT_START Test/[test]NN:OF:PLU:NOM:NEU|Test/[test]NN:OF:SIN:NOM:NEU|Test/[test]NN:OF:SIN:NOM:UTR  /[null]null 2/[null]null  /[null]null ./[...]<ELLIPS> ./[null]null ./[...]</ELLIPS>  /[null]null",
		},
	}
	for _, tc := range cases {
		got := myAssertSwedishChunker(tc.input, disambiguator)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}

	// Java: TestTools.myAssert(..., tokenizer, tagger) — no sentenceTokenizer, no disambiguator.
	// Format: word tokens only, readings joined by " -- ".
	gotVW := myAssertSwedishTaggerOnly("VW-skandalen tog fuskandet till en ny nivå.")
	wantVW := "VW-skandalen/[null]null -- tog/[ta]VB:PRT -- fuskandet/[null]null -- till/[till]AB|till/[till]PP -- en/[en]NN:OF:SIN:NOM:UTR|en/[en]PN -- ny/[ny]JJ:PU -- nivå/[null]null"
	require.Equal(t, wantVW, gotVW, "input=%q (tagger-only)", "VW-skandalen tog fuskandet till en ny nivå.")
}

// svMultiwordsPath resolves Java resource /sv/multiwords.txt under inspiration.
func svMultiwordsPath(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
	p := filepath.Join(dir,
		"inspiration/languagetool/languagetool-language-modules/sv/src/main/resources/org/languagetool/resource/sv/multiwords.txt")
	_, err = os.Stat(p)
	require.NoError(t, err, "Java /sv/multiwords.txt resource must exist")
	return p
}

// loadSwedishMultiWordChunker ports MultiWordChunker.getInstance("/sv/multiwords.txt")
// defaults: allowFirstCapitalized=false, allowAllUppercase=false, allowTitlecase=false.
// Prefer process-cached production wire (Java hybrid chunker field).
func loadSwedishMultiWordChunker(t *testing.T) *disambiguation.MultiWordChunker {
	t.Helper()
	if c := SwedishMultiWordChunker(); c != nil {
		return c
	}
	f, err := os.Open(svMultiwordsPath(t))
	require.NoError(t, err)
	defer f.Close()
	c, err := OpenSwedishMultiWordChunker(f)
	require.NoError(t, err)
	return c
}

// myAssertSwedishChunker ports Java TestTools.myAssert(input, expected, WordTokenizer,
// SRXSentenceTokenizer(Swedish), SwedishTagger, MultiWordChunker).
// Format: token/[lemma]POS readings sorted and joined by '|', tokens joined by space;
// null lemma/POS print as the literal "null" (Java string concat of null).
func myAssertSwedishChunker(input string, dis disambiguation.Disambiguator) string {
	tagsv.EnsureDefaultSwedishTagger()
	tagger := tagsv.DefaultSwedishTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("sv")
	var out strings.Builder
	for _, sentence := range st.Tokenize(input) {
		tokens := wt.Tokenize(sentence)
		var noWS []string
		for _, tok := range tokens {
			if swedishTestToolsIsWord(tok) {
				noWS = append(noWS, tok)
			}
		}
		aTokens := tagger.Tag(noWS)
		tokenArray := make([]*languagetool.AnalyzedTokenReadings, 0, len(tokens)+1)
		ss := languagetool.SentenceStartTagName
		tokenArray = append(tokenArray, languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken("", &ss, nil), 0))
		startPos := 0
		noWSCount := 0
		for _, tokenStr := range tokens {
			var posTag *languagetool.AnalyzedTokenReadings
			if swedishTestToolsIsWord(tokenStr) {
				posTag = aTokens[noWSCount]
				posTag.SetStartPos(startPos)
				noWSCount++
			} else {
				// Java BaseTagger.createNullToken / tagger.createNullToken
				posTag = languagetool.NewAnalyzedTokenReadingsAt(
					languagetool.NewAnalyzedToken(tokenStr, nil, nil), startPos)
			}
			tokenArray = append(tokenArray, posTag)
			startPos += tokenizers.UTF16Len(tokenStr)
		}
		finalSentence := languagetool.NewAnalyzedSentence(tokenArray)
		if dis != nil {
			finalSentence = dis.Disambiguate(finalSentence)
		}
		out.WriteString(formatSwedishMyAssertSentence(finalSentence))
	}
	return out.String()
}

// myAssertSwedishTaggerOnly ports Java TestTools.myAssert(input, expected, tokenizer, tagger):
// tokenize → drop non-word tokens → tag → sorted readings joined by " -- ".
func myAssertSwedishTaggerOnly(input string) string {
	tagsv.EnsureDefaultSwedishTagger()
	tagger := tagsv.DefaultSwedishTagger
	wt := tokenizers.NewWordTokenizer()
	tokens := wt.Tokenize(input)
	var noWS []string
	for _, tok := range tokens {
		if swedishTestToolsIsWord(tok) {
			noWS = append(noWS, tok)
		}
	}
	output := tagger.Tag(noWS)
	var parts []string
	for _, atr := range output {
		var readings []string
		for _, r := range atr.GetReadings() {
			if r != nil {
				readings = append(readings, swedishTestToolsGetAsString(r))
			}
		}
		sort.Strings(readings)
		parts = append(parts, strings.Join(readings, "|"))
	}
	return strings.Join(parts, " -- ")
}

// swedishTestToolsIsWord ports TestTools.isWord: any letter or digit → word token.
func swedishTestToolsIsWord(token string) bool {
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// formatSwedishMyAssertSentence ports TestTools.getAsStrings + join for one sentence.
func formatSwedishMyAssertSentence(sent *languagetool.AnalyzedSentence) string {
	if sent == nil {
		return ""
	}
	var parts []string
	for _, tr := range sent.GetTokens() {
		var readings []string
		for _, r := range tr.GetReadings() {
			if r != nil {
				readings = append(readings, swedishTestToolsGetAsString(r))
			}
		}
		// Java Collections.sort — force stable order across lexicon versions
		sort.Strings(readings)
		parts = append(parts, strings.Join(readings, "|"))
	}
	return strings.Join(parts, " ")
}

// swedishTestToolsGetAsString ports TestTools.getAsString: token/[lemma]POS with null literals.
func swedishTestToolsGetAsString(tok *languagetool.AnalyzedToken) string {
	lemma, pos := "null", "null"
	if tok.GetLemma() != nil {
		lemma = *tok.GetLemma()
	}
	if tok.GetPOSTag() != nil {
		pos = *tok.GetPOSTag()
	}
	return tok.GetToken() + "/[" + lemma + "]" + pos
}
