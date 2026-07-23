package disambiguation

// Twin of PolishDisambiguationRuleTest.testChunker
// Java: org.languagetool.tagging.disambiguation.PolishDisambiguationRuleTest#testChunker
// Resources: /pl/multiwords.txt, polish.dict via PolishTagger
import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	tagpl "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of PolishDisambiguationRuleTest.testChunker.
// Java: MultiWordChunker.getInstance("/pl/multiwords.txt") → false,false,false defaults;
// WordTokenizer (core), SRXSentenceTokenizer(Polish), PolishTagger; TestTools.myAssert.
func TestPolishDisambiguationRule_Chunker(t *testing.T) {
	if tagpl.DiscoverPolishPOSDict() == "" {
		t.Skip("polish.dict not in tree")
	}
	tagpl.EnsureDefaultPolishTagger()
	require.NotNil(t, tagpl.DefaultPolishTagger)
	require.NotNil(t, tagpl.DefaultPolishTagger.GetWordTagger())

	disambiguator := loadPolishMultiWordChunker(t)

	// Java TestTools.myAssert expected strings (active cases only; commented ones ignored).
	// Readings sorted like TestTools.getAsStrings.
	cases := []struct {
		input string
		want  string
	}{
		{
			"To test... dezambiguacji",
			"/[null]SENT_START To/[ten]adj:sg:acc:n1.n2:pos|To/[ten]adj:sg:nom.voc:n1.n2:pos|To/[to]conj|To/[to]qub|To/[to]subst:sg:acc:n2|To/[to]subst:sg:nom:n2  /[null]null test/[test]subst:sg:acc:m3|test/[test]subst:sg:nom:m3 ./[...]<ELLIPSIS> ./[null]null ./[...]</ELLIPSIS>  /[null]null dezambiguacji/[null]null",
		},
		{
			"On, to znaczy premier, jest niezbyt mądry",
			"/[null]SENT_START On/[on]adj:sg:acc:m3:pos|On/[on]adj:sg:nom.voc:m1.m2.m3:pos|On/[on]ppron3:sg:nom:m1.m2.m3:ter:akc.nakc:praep.npraep ,/[null]null  /[null]null to/[ten]adj:sg:acc:n1.n2:pos|to/[ten]adj:sg:nom.voc:n1.n2:pos|to/[to znaczy]<TO_ZNACZY>|to/[to]conj|to/[to]qub|to/[to]subst:sg:acc:n2|to/[to]subst:sg:nom:n2  /[null]null znaczy/[to znaczy]</TO_ZNACZY>|znaczy/[znaczyć]verb:fin:sg:ter:imperf:refl.nonrefl  /[null]null premier/[premier]subst:pl:acc:f|premier/[premier]subst:pl:dat:f|premier/[premier]subst:pl:gen:f|premier/[premier]subst:pl:inst:f|premier/[premier]subst:pl:loc:f|premier/[premier]subst:pl:nom:f|premier/[premier]subst:pl:voc:f|premier/[premier]subst:sg:acc:f|premier/[premier]subst:sg:dat:f|premier/[premier]subst:sg:gen:f|premier/[premier]subst:sg:inst:f|premier/[premier]subst:sg:loc:f|premier/[premier]subst:sg:nom:f|premier/[premier]subst:sg:nom:m1|premier/[premier]subst:sg:voc:f|premier/[premiera]subst:pl:gen:f ,/[null]null  /[null]null jest/[być]verb:fin:sg:ter:imperf:nonrefl  /[null]null niezbyt/[niezbyt]adv  /[null]null mądry/[mądry]adj:sg:acc:m3:pos|mądry/[mądry]adj:sg:nom.voc:m1.m2.m3:pos|mądry/[mądry]subst:sg:nom:m1|mądry/[mądry]subst:sg:voc:m1",
		},
		{
			"Lubię go z uwagi na krótkie włosy.",
			"/[null]SENT_START Lubię/[lubić]verb:fin:sg:pri:imperf:nonrefl|Lubię/[lubić]verb:fin:sg:pri:imperf:refl.nonrefl  /[null]null go/[go]subst:pl:acc:n2|go/[go]subst:pl:dat:n2|go/[go]subst:pl:gen:n2|go/[go]subst:pl:inst:n2|go/[go]subst:pl:loc:n2|go/[go]subst:pl:nom:n2|go/[go]subst:pl:voc:n2|go/[go]subst:sg:acc:n2|go/[go]subst:sg:dat:n2|go/[go]subst:sg:gen:n2|go/[go]subst:sg:inst:n2|go/[go]subst:sg:loc:n2|go/[go]subst:sg:nom:n2|go/[go]subst:sg:voc:n2|go/[on]ppron3:sg:acc:m1.m2.m3:ter:nakc:npraep|go/[on]ppron3:sg:gen:m1.m2.m3:ter:nakc:npraep|go/[on]ppron3:sg:gen:n1.n2:ter:nakc:npraep  /[null]null z/[z uwagi na]<PREP:ACC>|z/[z]prep:acc:nwok|z/[z]prep:gen:nwok|z/[z]prep:inst:nwok  /[null]null uwagi/[uwaga]subst:pl:acc:f|uwagi/[uwaga]subst:pl:nom:f|uwagi/[uwaga]subst:pl:voc:f|uwagi/[uwaga]subst:sg:gen:f  /[null]null na/[na]interj|na/[na]prep:acc|na/[na]prep:loc|na/[z uwagi na]</PREP:ACC>  /[null]null krótkie/[krótki]adj:pl:acc:m2.m3.f.n1.n2.p2.p3:pos|krótkie/[krótki]adj:pl:nom.voc:m2.m3.f.n1.n2.p2.p3:pos|krótkie/[krótki]adj:sg:acc:n1.n2:pos|krótkie/[krótki]adj:sg:nom.voc:n1.n2:pos  /[null]null włosy/[włos]subst:pl:acc:m3|włosy/[włos]subst:pl:nom:m3|włosy/[włos]subst:pl:voc:m3|włosy/[włosy]subst:pl:acc:p3|włosy/[włosy]subst:pl:nom:p3|włosy/[włosy]subst:pl:voc:p3 ./[null]null",
		},
		{
			"Test...",
			"/[null]SENT_START Test/[test]subst:sg:acc:m3|Test/[test]subst:sg:nom:m3 ./[...]<ELLIPSIS> ./[null]null ./[...]</ELLIPSIS>",
		},
		{
			"Test... ",
			"/[null]SENT_START Test/[test]subst:sg:acc:m3|Test/[test]subst:sg:nom:m3 ./[...]<ELLIPSIS> ./[null]null ./[...]</ELLIPSIS>  /[null]null",
		},
	}
	for _, tc := range cases {
		got := myAssertPolishChunker(tc.input, disambiguator)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// plMultiwordsPath resolves Java resource /pl/multiwords.txt under inspiration.
func plMultiwordsPath(t *testing.T) string {
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
		"inspiration/languagetool/languagetool-language-modules/pl/src/main/resources/org/languagetool/resource/pl/multiwords.txt")
	_, err = os.Stat(p)
	require.NoError(t, err, "Java /pl/multiwords.txt resource must exist")
	return p
}

// loadPolishMultiWordChunker ports MultiWordChunker.getInstance("/pl/multiwords.txt")
// defaults: allowFirstCapitalized=false, allowAllUppercase=false, allowTitlecase=false.
func loadPolishMultiWordChunker(t *testing.T) *MultiWordChunker {
	t.Helper()
	f, err := os.Open(plMultiwordsPath(t))
	require.NoError(t, err)
	defer f.Close()
	c, err := NewMultiWordChunkerFromReader(f, MultiWordChunkerSettings{
		AllowFirstCapitalized: false,
		AllowAllUppercase:     false,
		AllowTitlecase:        false,
	})
	require.NoError(t, err)
	return c
}

// myAssertPolishChunker ports Java TestTools.myAssert(input, expected, WordTokenizer,
// SRXSentenceTokenizer(Polish), PolishTagger, MultiWordChunker).
// Format: token/[lemma]POS readings sorted and joined by '|', tokens joined by space;
// null lemma/POS print as the literal "null" (Java string concat of null).
func myAssertPolishChunker(input string, dis Disambiguator) string {
	tagpl.EnsureDefaultPolishTagger()
	tagger := tagpl.DefaultPolishTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("pl")
	var out strings.Builder
	for _, sentence := range st.Tokenize(input) {
		tokens := wt.Tokenize(sentence)
		var noWS []string
		for _, tok := range tokens {
			if polishTestToolsIsWord(tok) {
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
			if polishTestToolsIsWord(tokenStr) {
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
		out.WriteString(formatPolishMyAssertSentence(finalSentence))
	}
	return out.String()
}

// polishTestToolsIsWord ports TestTools.isWord: any letter or digit → word token.
func polishTestToolsIsWord(token string) bool {
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// formatPolishMyAssertSentence ports TestTools.getAsStrings + join for one sentence.
func formatPolishMyAssertSentence(sent *languagetool.AnalyzedSentence) string {
	if sent == nil {
		return ""
	}
	var parts []string
	for _, tr := range sent.GetTokens() {
		var readings []string
		for _, r := range tr.GetReadings() {
			if r != nil {
				readings = append(readings, polishTestToolsGetAsString(r))
			}
		}
		// Java Collections.sort — force stable order across lexicon versions
		sort.Strings(readings)
		parts = append(parts, strings.Join(readings, "|"))
	}
	return strings.Join(parts, " ")
}

// polishTestToolsGetAsString ports TestTools.getAsString: token/[lemma]POS with null literals.
func polishTestToolsGetAsString(tok *languagetool.AnalyzedToken) string {
	lemma, pos := "null", "null"
	if tok.GetLemma() != nil {
		lemma = *tok.GetLemma()
	}
	if tok.GetPOSTag() != nil {
		pos = *tok.GetPOSTag()
	}
	return tok.GetToken() + "/[" + lemma + "]" + pos
}
