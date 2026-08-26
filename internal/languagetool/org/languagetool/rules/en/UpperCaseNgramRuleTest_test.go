package en

// Twin of org.languagetool.rules.en.UpperCaseNgramRuleTest.

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/languagemodel"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	tagen "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	"github.com/stretchr/testify/require"
)

// fakeLanguageModelFromJavaMap ports FakeLanguageModel(map) counts used by UpperCaseNgramRuleTest.
// Keys are space-joined ngrams (Java String.join(" ", tokens)); total = sum of all map values.
func fakeLanguageModelFromJavaMap(m map[string]int) ngrams.LanguageModel {
	lm := languagemodel.NewMapLanguageModel()
	var total int64
	for k, v := range m {
		c := int64(v)
		parts := strings.Split(k, " ")
		lm.Add(parts, c)
		total += c
	}
	lm.Total = total
	return lm
}

func newUpperCaseNgramRuleForTest() *UpperCaseNgramRule {
	// Java UpperCaseNgramRuleTest static map
	m := map[string]int{
		"really like":  100,
		"like spaghetti": 100,
		"This was":     100,
		"This was a":   10,
		"this was":     100,
		"this was a":   10,
		"indeed was":   100,
		"indeed was a": 10,
	}
	return NewUpperCaseNgramRule(nil, fakeLanguageModelFromJavaMap(m), "en", nil)
}

func TestUpperCaseNgramRule_Rule(t *testing.T) {
	rule := newUpperCaseNgramRuleForTest()
	require.Equal(t, "EN_UPPER_CASE_NGRAM", rule.GetID())

	assertGood := func(s string) {
		t.Helper()
		sent := tagen.AnalyzeEnglishSentence(s)
		require.NotNil(t, sent, "analyze %q", s)
		ms, err := rule.Match(sent)
		require.NoError(t, err)
		require.Empty(t, ms, "Expected no matches for %q, got %#v", s, ms)
	}
	assertMatch := func(s string) {
		t.Helper()
		sent := tagen.AnalyzeEnglishSentence(s)
		require.NotNil(t, sent, "analyze %q", s)
		ms, err := rule.Match(sent)
		require.NoError(t, err)
		require.Len(t, ms, 1, "Expected 1 match for %q, got %#v", s, ms)
	}

	// Java assertGood matrix
	assertGood("The New York Times reviews their gallery all the time.") // spelling_global
	assertGood("This Was a Good Idea")                                   // no dot = no real sentence
	assertGood("Professor Sprout acclimated the plant to a new environment.")
	assertGood("Beauty products, Clean & Clear facial wash.")
	assertGood("Please click Account > Withdraw > Update.")
	assertGood("The goal is to Develop, Discuss and Learn.")
	assertGood("(b) Summarize the strategy.")
	assertGood("Figure/Ground:")
	assertGood("What Happened?")
	assertGood("1- Have you personally made any improvements?")
	assertGood("Lesson #1 - Create a webinar.")
	assertGood("Please refund Order #5698656.")
	assertGood("Let's play games at Games.co.uk.")
	assertGood("Ben (Been).")
	assertGood("C stands for Curse.")
	assertGood("The United States also used the short-lived slogan, \"Tastes So Good, You'll Roar\", in the early 1980s.")
	assertGood("09/06 - Spoken to the business manager.")
	assertGood("12.3 Game.")
	assertGood("Let's talk to the Onboarding team.")
	assertGood("My name is Gentle.")
	assertGood("They called it Greet.")
	assertGood("What is Foreshadowing?")
	assertGood("His name is Carp.")
	assertGood("Victor or Rabbit as everyone calls him.")
	assertGood("Think I'm Tripping?")
	assertGood("Music and Concepts.")
	assertGood("It is called Ranked mode.")
	assertGood("I was into Chronicle of a Death Foretold.")
	assertGood("I talked with Engineering.")
	assertGood("They used Draft.js to solve it.")
	assertGood("And mine is Wed.")
	assertGood("I would support Knicks rather than Hawks.")
	assertGood("You Can't Judge a Book by the Cover")
	assertGood("What Does an Effective Cover Letter Look Like?")
	assertGood("Our external Counsel are reviewing the authority of FMPA to enter into the proposed transaction")
	assertGood("Otherwise, Staff will proceed to process your filing based on the pro forma tariff sheets submitted on August 15, 2000.")
	assertGood("But he is not accomplishing enough statistically to help most Fantasy teams.")

	// Parentheses
	assertGood("(4 hrs/wk) Manage all IT affairs.")
	assertGood("(Laravel MVC) Implements two distinct working algorithms.")
	assertGood("(Later) Connect different cont.")

	assertGood("$$/month (Includes everything!)")
	assertGood("- Foot care (Cleaning of feet, wash...")
	assertGood("- Exercise (Engage in exercises...")

	// Arrows
	assertGood("-> Allowed the civilian government...")
	assertGood("-> Led by Italian Physicians...")
	assertGood("-> Used as inspiration...")

	// Quotes
	assertGood("The sign read \"Seats to be Added\".")
	assertGood("“Helm, Engage.”")
	assertGood("\"Be careful, Reign!\"")

	assertGood("ii) Expanded the notes.")

	// Java assertMatch
	assertMatch("I really Like spaghetti.")
	assertMatch("This Was a good idea.")
	assertMatch("But this Was a good idea.")
	assertMatch("This indeed Was a good idea.")
}

func TestUpperCaseNgramRule_FirstLongWordToLeftIsUppercase(t *testing.T) {
	rule := newUpperCaseNgramRuleForTest()

	// Java: lt.getAnalyzedSentence(...).getTokens() — includes whitespace tokens
	sent1 := tagen.AnalyzeEnglishSentence("As with Lifeboat and Rope, the principal characters were ...")
	require.NotNil(t, sent1)
	tokens1 := sent1.GetTokens()
	// left:
	require.False(t, rule.FirstLongWordToLeftIsUppercase(tokens1, 1)) // 1 = "As"
	require.True(t, rule.FirstLongWordToLeftIsUppercase(tokens1, 9))  // 9 = "Rope"
	// right:
	require.True(t, rule.FirstLongWordToRightIsUppercase(tokens1, 3))   // 3 = "with"
	require.True(t, rule.FirstLongWordToRightIsUppercase(tokens1, 5))   // 5 = "Lifeboat"
	require.False(t, rule.FirstLongWordToRightIsUppercase(tokens1, 10)) // 10 = ","

	sent2 := tagen.AnalyzeEnglishSentence("From Theory to Practice, followed by some other words")
	require.NotNil(t, sent2)
	tokens2 := sent2.GetTokens()
	// left:
	require.False(t, rule.FirstLongWordToLeftIsUppercase(tokens2, 3))  // 3 = "Theory"
	require.True(t, rule.FirstLongWordToLeftIsUppercase(tokens2, 4))   // 4 = "to" (ws layout: looks left → Theory)
	require.True(t, rule.FirstLongWordToLeftIsUppercase(tokens2, 7))   // 7 = "Practice"
	require.False(t, rule.FirstLongWordToLeftIsUppercase(tokens2, 1))  // 1 = "From" (sentence start)
	require.False(t, rule.FirstLongWordToLeftIsUppercase(tokens2, 12)) // 12 = "by"
	// right:
	require.False(t, rule.FirstLongWordToRightIsUppercase(tokens2, 8))  // 8 = ","
	require.False(t, rule.FirstLongWordToRightIsUppercase(tokens2, 10)) // 10 = "followed"
}

func TestUpperCaseNgramRule_AntiPatternsCount(t *testing.T) {
	// Java ANTI_PATTERNS has 85 entries
	require.Equal(t, 85, len(UpperCaseNgramAntiPatterns), "Java ANTI_PATTERNS 85/85")
}
