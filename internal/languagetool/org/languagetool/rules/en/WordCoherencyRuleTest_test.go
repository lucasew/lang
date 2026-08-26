package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/WordCoherencyRuleTest.java
//
// Java setup:
//   JLanguageTool lt = new JLanguageTool(Languages.getLanguageForShortCode("en-US"));
//   TestTools.disableAllRulesExcept(lt, "EN_WORD_COHERENCY");
//   assertGood/assertError → rule.match(lt.analyzeText(s))
//   positions/suggestions → lt.check(s) with only EN_WORD_COHERENCY enabled
//
// Go analysis path (same morph lemmas Java uses):
//   SRXSentenceTokenizer("en-US") + tagging/en.AnalyzeEnglishSentence per part
//   (EnglishWordTokenizer + EnglishTagger english.dict + hybrid disambiguator).
// Production loads coherency.txt pairs only (no invent -ed/-ing expand).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	tagen "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

func engWordCoherencyLT(t *testing.T) {
	t.Helper()
	if tagen.DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree (third_party/english-pos-dict or inspiration)")
	}
	tagen.EnsureDefaultEnglishTagger()
	require.NotNil(t, tagen.EnglishTagWord(), "EnglishTagWord must load from english.dict")
}

// analyzeENWordCoherencyText ports JLanguageTool(en-US).analyzeText / check:
// sentence tokenize then getAnalyzedSentence per part (English tagger lemmas).
func analyzeENWordCoherencyText(t *testing.T, text string) []*languagetool.AnalyzedSentence {
	t.Helper()
	parts := tokenizers.NewSRXSentenceTokenizer("en-US").Tokenize(text)
	out := make([]*languagetool.AnalyzedSentence, 0, len(parts))
	for _, p := range parts {
		s := tagen.AnalyzeEnglishSentence(p)
		require.NotNil(t, s, "AnalyzeEnglishSentence(%q)", p)
		out = append(out, s)
	}
	return out
}

// TestWordCoherencyRule_Rule twins Java WordCoherencyRuleTest.testRule.
func TestWordCoherencyRule_Rule(t *testing.T) {
	engWordCoherencyLT(t)
	rule := NewWordCoherencyRule(nil)

	assertGood := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.MatchList(analyzeENWordCoherencyText(t, s))), "good: %q", s)
	}
	assertError := func(s string) {
		t.Helper()
		require.Equal(t, 1, len(rule.MatchList(analyzeENWordCoherencyText(t, s))), "error: %q", s)
	}

	// correct sentences:
	assertGood("He likes archeology. She likes archeology, too.")
	assertGood("He likes archaeology. She likes archaeology, too.")
	// errors:
	assertError("He likes archaeology. She likes archeology, too.")

	matches1 := rule.MatchList(analyzeENWordCoherencyText(t, "He is reelected, or he will be re-elected."))
	require.Equal(t, 1, len(matches1))
	require.Equal(t, 31, matches1[0].GetFromPos())
	require.Equal(t, 41, matches1[0].GetToPos())
	require.Equal(t, []string{"reelected"}, matches1[0].GetSuggestedReplacements())

	matches2 := rule.MatchList(analyzeENWordCoherencyText(t, "He was reelected, and I will re-elect him again in 2002."))
	require.Equal(t, 1, len(matches2))
	require.Equal(t, 29, matches2[0].GetFromPos())
	require.Equal(t, 37, matches2[0].GetToPos())
	require.Equal(t, []string{"reelect"}, matches2[0].GetSuggestedReplacements())

	matches3 := rule.MatchList(analyzeENWordCoherencyText(t, "He oxidises o, or he oxidizes"))
	require.Equal(t, 1, len(matches3))
	require.Equal(t, 21, matches3[0].GetFromPos())
	require.Equal(t, 29, matches3[0].GetToPos())
	require.Equal(t, []string{"oxidises"}, matches3[0].GetSuggestedReplacements())
}

// TestWordCoherencyRule_CallIndependence twins Java testCallIndependence.
func TestWordCoherencyRule_CallIndependence(t *testing.T) {
	engWordCoherencyLT(t)
	// Separate rule instances / match calls do not share state
	assertGood := func(s string) {
		t.Helper()
		rule := NewWordCoherencyRule(nil)
		require.Equal(t, 0, len(rule.MatchList(analyzeENWordCoherencyText(t, s))))
	}
	assertGood("He likes archaeology.")
	assertGood("She likes archeology, too.") // not noticed across independent calls
}

// TestWordCoherencyRule_Metadata twins Java WordCoherencyRule id/description/example pair.
func TestWordCoherencyRule_Metadata(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	require.Equal(t, "EN_WORD_COHERENCY", rule.GetID())
	require.Equal(t, "Coherent spelling of words with two admitted variants.", rule.GetDescription())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "He likes archaeology. Really? She likes <marker>archeology</marker>, too.", inc[0].GetExample())
	require.Equal(t, []string{"archaeology"}, inc[0].GetCorrections())
	require.Equal(t, "He likes archaeology. Really? She likes <marker>archaeology</marker>, too.", rule.GetCorrectExamples()[0].GetExample())
}

// TestWordCoherencyRule_MatchPosition twins Java testMatchPosition.
func TestWordCoherencyRule_MatchPosition(t *testing.T) {
	engWordCoherencyLT(t)
	rule := NewWordCoherencyRule(nil)
	matches := rule.MatchList(analyzeENWordCoherencyText(t, "He likes archaeology. She likes archeology, too."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 32, matches[0].GetFromPos())
	require.Equal(t, 42, matches[0].GetToPos())
}

// TestWordCoherencyRule_RuleCompleteTexts twins Java testRuleCompleteTexts.
func TestWordCoherencyRule_RuleCompleteTexts(t *testing.T) {
	engWordCoherencyLT(t)
	// Java uses lt.check with only EN_WORD_COHERENCY; we run WordCoherencyRule alone.
	check := func(s string) int {
		return len(NewWordCoherencyRule(nil).MatchList(analyzeENWordCoherencyText(t, s)))
	}
	require.Equal(t, 0, check("He likes archaeology. Really? She likes archaeology, too."))
	require.Equal(t, 1, check("He likes archaeology. Really? She likes archeology, too."))
	require.Equal(t, 1, check("He likes archeology. Really? She likes archaeology, too."))
	require.Equal(t, 1, check("Mix of upper case and lower case: Westernize and westernise."))
}
