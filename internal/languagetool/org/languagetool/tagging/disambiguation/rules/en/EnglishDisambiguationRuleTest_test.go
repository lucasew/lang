package en

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambigxx "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/xx"
	tagen "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of EnglishDisambiguationRuleTest.testChunker (Java TestTools.myAssert strings).
// Java: org.languagetool.tagging.disambiguation.rules.en.EnglishDisambiguationRuleTest#testChunker
func TestEnglishDisambiguationRule_Chunker(t *testing.T) {
	if tagen.DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree")
	}
	tagen.EnsureDefaultEnglishTagger()
	require.NotNil(t, tagen.DefaultEnglishTagger)
	require.NotNil(t, tagen.DefaultEnglishTagger.GetWordTagger())

	xmlDisam := tagen.EnglishXmlRuleDisambiguator()
	if xmlDisam == nil || len(xmlDisam.Rules) == 0 {
		t.Skip("en/disambiguation.xml not in tree")
	}
	demoDisam := disambigxx.NewDemoDisambiguator()
	hybridDisam := tagen.DefaultEnglishHybridDisambiguator()
	if hybridDisam == nil || hybridDisam.Chunker == nil {
		t.Skip("en multiwords / hybrid resources not in tree")
	}

	// Java TestTools.myAssert expected strings (order of readings is sorted in TestTools).
	cases := []struct {
		input string
		want  string
		dis   disambiguation.Disambiguator
	}{
		{
			"I cannot have it.",
			"/[null]SENT_START I/[I]PRP|I/[I]PRP_S1S  /[null]null cannot/[can]MD  /[null]null have/[have]VB  /[null]null it/[it]PRP|it/[it]PRP_O3SN|it/[it]PRP_S3SN ./[.]PCT",
			xmlDisam,
		},
		{
			"I cannot have it.",
			"/[null]SENT_START I/[I]PRP|I/[I]PRP_S1S  /[null]null cannot/[can]MD  /[null]null have/[have]NN|have/[have]VB|have/[have]VBP  /[null]null it/[it]PRP|it/[it]PRP_O3SN|it/[it]PRP_S3SN ./[null]null",
			demoDisam,
		},
		{
			"He is to blame.",
			"/[null]SENT_START He/[he]PRP|He/[he]PRP_S3SM  /[null]null is/[be]VBZ  /[null]null to/[to]IN|to/[to]TO  /[null]null blame/[blame]VB ./[.]PCT",
			xmlDisam,
		},
		{
			"He is to blame.",
			"/[null]SENT_START He/[he]PRP|He/[he]PRP_S3SM  /[null]null is/[be]VBZ  /[null]null to/[to]IN|to/[to]TO  /[null]null blame/[blame]JJ|blame/[blame]NN:UN|blame/[blame]VB|blame/[blame]VBP ./[null]null",
			demoDisam,
		},
		{
			"He is well known.",
			"/[null]SENT_START He/[he]PRP|He/[he]PRP_S3SM  /[null]null is/[be]VBZ  /[null]null well/[well]JJ|well/[well]NN|well/[well]RB|well/[well]UH|well/[well]VB|well/[well]VBP  /[null]null known/[know]VBN|known/[known]NN ./[null]null",
			demoDisam,
		},
		{
			"The quid pro quo.",
			"/[null]SENT_START The/[the]DT  /[null]null quid/[quid pro quo]NN  /[null]null pro/[quid pro quo]NN  /[null]null quo/[quid pro quo]NN ./[.]PCT",
			hybridDisam,
		},
		{
			"The QUID PRO QUO.",
			"/[null]SENT_START The/[the]DT  /[null]null QUID/[quid pro quo]NN  /[null]null PRO/[quid pro quo]NN  /[null]null QUO/[quid pro quo]NN ./[.]PCT",
			hybridDisam,
		},
	}
	for _, tc := range cases {
		got := myAssertDisambiguate(tc.input, tc.dis)
		require.Equal(t, tc.want, got, "input=%q dis=%T", tc.input, tc.dis)
	}
}

// Twin of disambiguation.xml QUARAN example: Qur'an → an[NNP]
func TestEnglishDisambiguationRule_QuranAnNNP(t *testing.T) {
	if tagen.DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree")
	}
	sent := tagen.AnalyzeEnglishSentence("Qur'an.")
	an := findTok(sent, "an")
	require.NotNil(t, an)
	require.True(t, an.IsIgnoredBySpeller(), "multiword ignore-spelling")
	require.True(t, hasPOS(an, "NNP"), "QUARAN / multiword NNP: %s", dumpTags(an))
}

// Twin of UNKNOWN_PCT via hybrid XML path (also covered in tagging/en).
func TestEnglishDisambiguationRule_UnknownPCT(t *testing.T) {
	if tagen.DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree")
	}
	sent := tagen.AnalyzeEnglishSentence("Hello.")
	dot := findTok(sent, ".")
	require.NotNil(t, dot)
	require.True(t, hasPOS(dot, "PCT"), dumpTags(dot))
}

// myAssertDisambiguate ports Java TestTools.myAssert(input, expected, WordTokenizer,
// SRXSentenceTokenizer, EnglishTagger, disambiguator) for EN disambiguation twins.
// Format: token/[lemma]POS readings sorted and joined by '|', tokens joined by space;
// null lemma/POS print as the literal "null" (Java string concat of null).
func myAssertDisambiguate(input string, dis disambiguation.Disambiguator) string {
	tagen.EnsureDefaultEnglishTagger()
	tagger := tagen.DefaultEnglishTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("en")
	var out strings.Builder
	for _, sentence := range st.Tokenize(input) {
		tokens := wt.Tokenize(sentence)
		var noWS []string
		for _, tok := range tokens {
			if testToolsIsWord(tok) {
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
			if testToolsIsWord(tokenStr) {
				posTag = aTokens[noWSCount]
				posTag.SetStartPos(startPos)
				noWSCount++
			} else {
				// Java BaseTagger.createNullToken
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
		out.WriteString(formatMyAssertSentence(finalSentence))
	}
	return out.String()
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

// formatMyAssertSentence ports TestTools.getAsStrings + join for one sentence.
func formatMyAssertSentence(sent *languagetool.AnalyzedSentence) string {
	if sent == nil {
		return ""
	}
	var parts []string
	for _, tr := range sent.GetTokens() {
		var readings []string
		for _, r := range tr.GetReadings() {
			if r != nil {
				readings = append(readings, testToolsGetAsString(r))
			}
		}
		// Java Collections.sort — force stable order across lexicon versions
		sort.Strings(readings)
		parts = append(parts, strings.Join(readings, "|"))
	}
	return strings.Join(parts, " ")
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

func findTok(sent *languagetool.AnalyzedSentence, surface string) *languagetool.AnalyzedTokenReadings {
	if sent == nil {
		return nil
	}
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == surface {
			return tok
		}
	}
	return nil
}

func hasPOS(tok *languagetool.AnalyzedTokenReadings, pos string) bool {
	if tok == nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r != nil && r.GetPOSTag() != nil && *r.GetPOSTag() == pos {
			return true
		}
	}
	return false
}

func dumpTags(tok *languagetool.AnalyzedTokenReadings) string {
	if tok == nil {
		return ""
	}
	var b strings.Builder
	for _, r := range tok.GetReadings() {
		if r == nil {
			continue
		}
		p, l := "", ""
		if r.GetPOSTag() != nil {
			p = *r.GetPOSTag()
		}
		if r.GetLemma() != nil {
			l = *r.GetLemma()
		}
		if b.Len() > 0 {
			b.WriteByte('|')
		}
		b.WriteString(l)
		b.WriteByte('/')
		b.WriteString(p)
	}
	return b.String()
}
