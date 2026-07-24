package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishUnpairedBracketsRuleTest.java
//
// Java setup:
//   rule = new EnglishUnpairedBracketsRule(TestTools.getEnglishMessages(), Languages.getLanguageForShortCode("en"));
//   lt   = new JLanguageTool(Languages.getLanguageForShortCode("en"));
// assertCorrect/Incorrect: rule.match(singletonList(lt.getAnalyzedSentence(s)))
// assertCorrectText:       rule.match(lt.analyzeText(s), annotatedText)
// testMultipleSentences:   lt.check(input).size() with only unpaired-brackets relevant counts
//                          → here MatchList(lt.Analyze) isolates the TextLevelRule path
//                          (full English JLT Check would mix other rules).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func engMessagesEUBR() map[string]string {
	// MessagesBundle_en keys used by GenericUnpairedBracketsRule / category.
	return map[string]string{
		"unpaired_brackets":      "Unpaired symbol: ''{0}'' seems to be missing",
		"desc_unpaired_brackets": "Unpaired braces, brackets, quotation marks and similar symbols",
		"category_punctuation":   "Punctuation",
	}
}

func setUpEUBR() (*EnglishUnpairedBracketsRule, *languagetool.JLanguageTool) {
	lt := languagetool.NewJLanguageTool("en")
	rule := NewEnglishUnpairedBracketsRule(engMessagesEUBR())
	return rule, lt
}

// assertCorrect ports assertCorrect: rule.match(singletonList(lt.getAnalyzedSentence(sentence))).
func assertCorrectEUBR(t *testing.T, rule *EnglishUnpairedBracketsRule, lt *languagetool.JLanguageTool, sentence string) {
	t.Helper()
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{lt.GetAnalyzedSentence(sentence)})
	require.Equal(t, 0, len(ms), "correct %q", sentence)
}

// assertIncorrect ports assertIncorrect: expect exactly one match.
func assertIncorrectEUBR(t *testing.T, rule *EnglishUnpairedBracketsRule, lt *languagetool.JLanguageTool, sentence string) {
	t.Helper()
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{lt.GetAnalyzedSentence(sentence)})
	require.Equal(t, 1, len(ms), "incorrect %q", sentence)
}

// assertCorrectText ports assertCorrectText: rule.match(lt.analyzeText(sentences), aText).
// AnnotatedText is unused by GenericUnpairedBracketsRule.match(List).
func assertCorrectTextEUBR(t *testing.T, rule *EnglishUnpairedBracketsRule, lt *languagetool.JLanguageTool, sentences string) {
	t.Helper()
	ms := rule.MatchList(lt.Analyze(sentences))
	require.Equal(t, 0, len(ms), "correct text %q", sentences)
}

func TestEnglishUnpairedBracketsRule_Rule(t *testing.T) {
	rule, lt := setUpEUBR()

	// correct sentences (Java assertCorrect):
	assertCorrectEUBR(t, rule, lt, "(This is a test sentence).")
	assertCorrectEUBR(t, rule, lt, "This is a word 'test'.")
	assertCorrectEUBR(t, rule, lt, "This is no smiley: (some more text)")
	assertCorrectEUBR(t, rule, lt, "This is a sentence with a smiley :)")
	assertCorrectEUBR(t, rule, lt, "This is a sentence with a smiley :(")
	assertCorrectEUBR(t, rule, lt, "This is a sentence with a smiley :-)")
	assertCorrectEUBR(t, rule, lt, "This is a sentence with a smiley ;-) and so on...")
	assertCorrectEUBR(t, rule, lt, "This is a [test] sentence...")
	assertCorrectEUBR(t, rule, lt, "The plight of Tamil refugees caused a surge of support from most of the Tamil political parties.[90]")
	assertCorrectEUBR(t, rule, lt, "(([20] [20] [20]))")

	// false alarm after disambiguation (quotes not in this rule)
	assertCorrectEUBR(t, rule, lt, "This is a \"special test\", right?")
	// numerical bullets / chapter refs
	assertCorrectEUBR(t, rule, lt, "We discussed this in Chapter 1).")
	assertCorrectEUBR(t, rule, lt, "The jury recommended that: (1) Four additional deputies be employed.")
	assertCorrectEUBR(t, rule, lt, "We discussed this in section 1a).")
	assertCorrectEUBR(t, rule, lt, "We discussed this in section iv).")
	// more cases
	assertCorrectEUBR(t, rule, lt, "(Ketab fi Isti'mal al-'Adad al-Hindi)")
	assertCorrectEUBR(t, rule, lt, "(al-'Adad al-Hindi)")
	assertCorrectEUBR(t, rule, lt, "will-o'-the-wisp")
	assertCorrectEUBR(t, rule, lt, "cat-o’-nine-tails")
	// Java TODO: ambiguous — still assertCorrect
	assertCorrectEUBR(t, rule, lt, "a) item one\nb) item two\nc) item three")
	// Java assertCorrectText (analyzeText multi-sentence path)
	assertCorrectTextEUBR(t, rule, lt, "\n\n"+
		"a) New York\n"+
		"b) Boston\n")
	assertCorrectTextEUBR(t, rule, lt, "\n\n"+
		"1.) New York\n"+
		"2.) Boston\n")
	assertCorrectTextEUBR(t, rule, lt, "\n\n"+
		"XII.) New York\n"+
		"XIII.) Boston\n")
	assertCorrectTextEUBR(t, rule, lt, "\n\n"+
		"A) New York\n"+
		"B) Boston\n"+
		"C) Foo\n")

	// incorrect sentences:
	assertIncorrectEUBR(t, rule, lt, "(This is a test sentence.")
	// unfinished paren without sentence-end punctuation → 0 (Java assertCorrect)
	assertCorrectEUBR(t, rule, lt, "This is not so (neither a nor b")
	assertIncorrectEUBR(t, rule, lt, "This is not so (neither a nor b.")
	assertIncorrectEUBR(t, rule, lt, "This is not so neither a nor b)")
	assertIncorrectEUBR(t, rule, lt, "This is not so neither foo nor bar)")

	// smiley inside parens: no ! → correct; with ! → incorrect (Java)
	assertCorrectEUBR(t, rule, lt, "Some text (and some funny remark :-) with more text to follow")
	assertIncorrectEUBR(t, rule, lt, "Some text (and some funny remark :-) with more text to follow!")

	// multi-match (mismatched bracket types)
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{lt.GetAnalyzedSentence("(This is a test] sentence.")})
	require.Equal(t, 2, len(ms))
	ms = rule.MatchList([]*languagetool.AnalyzedSentence{lt.GetAnalyzedSentence("This [is (a test} sentence.")})
	require.Equal(t, 3, len(ms))
}

func TestEnglishUnpairedBracketsRule_MultipleSentences(t *testing.T) {
	// Java testMultipleSentences: getMatches → lt.check(input).size()
	// Isolate EN_UNPAIRED_BRACKETS (full en Check mixes other rules).
	rule, lt := setUpEUBR()
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		lt.DisableRule(id)
	}
	lt.AddTextLevelRuleChecker(rule.GetID(), rules.AsTextLevelChecker(rule.MatchList))

	// also direct TextLevelRule path: rule.match(lt.analyzeText(...))
	matchN := func(input string) int {
		return len(rule.MatchList(lt.Analyze(input)))
	}
	checkN := func(input string) int {
		return len(lt.Check(input))
	}

	in0a := "This is multiple sentence text that contains a bracket: " +
		"[This is a bracket. With some text.] and this continues.\n"
	require.Equal(t, 0, matchN(in0a))
	require.Equal(t, 0, checkN(in0a))

	in0b := "This is multiple sentence text that contains a bracket. " +
		"(This is a bracket. \n\n With some text.) and this continues."
	require.Equal(t, 0, matchN(in0b))
	require.Equal(t, 0, checkN(in0b))

	in1 := "This is multiple sentence text that contains a bracket: " +
		"[This is a bracket. With some text. And this continues.\n\n"
	require.Equal(t, 1, matchN(in1))
	require.Equal(t, 1, checkN(in1))
}

// Java EnglishUnpairedBracketsRule: EN_UNPAIRED_BRACKETS, parentheses URL, example pair.
func TestEnglishUnpairedBracketsRule_Metadata(t *testing.T) {
	rule := NewEnglishUnpairedBracketsRule(engMessagesEUBR())
	require.Equal(t, "EN_UNPAIRED_BRACKETS", rule.GetID())
	require.Contains(t, rule.GetURL(), "what-are-parentheses")
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "PUNCTUATION", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSTypographical, rule.GetLocQualityIssueType())
	require.Equal(t, engMessagesEUBR()["desc_unpaired_brackets"], rule.GetDescription())
	require.Equal(t, -1, rule.MinToCheckParagraph())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "He lived in a <marker>(</marker>large house.", inc[0].GetExample())
	// Java Rule.addExamplePair uses first fixed <marker> span as correction
	require.Equal(t, []string{"("}, inc[0].GetCorrections())
	require.Contains(t, rule.GetCorrectExamples()[0].GetExample(), "large")
	// MatchList rebinds Rule so match.Rule is the EN type with EN id
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.NewJLanguageTool("en").GetAnalyzedSentence("(open only."),
	})
	require.NotEmpty(t, ms)
	require.Equal(t, rule, ms[0].Rule)
}
