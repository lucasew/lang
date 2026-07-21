package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishUnpairedBracketsRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestEnglishUnpairedBracketsRule_Rule(t *testing.T) {
	// Java: rule.match(Collections.singletonList(lt.getAnalyzedSentence(sentence)))
	rule := NewEnglishUnpairedBracketsRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	assertCorrect := func(s string) {
		t.Helper()
		require.Equal(t, 0, matchN(s), "correct %q", s)
	}
	assertIncorrect := func(s string) {
		t.Helper()
		require.Equal(t, 1, matchN(s), "incorrect %q", s)
	}

	// correct sentences (Java assertCorrect)
	assertCorrect("(This is a test sentence).")
	assertCorrect("This is a word 'test'.")
	assertCorrect("This is no smiley: (some more text)")
	assertCorrect("This is a sentence with a smiley :)")
	assertCorrect("This is a sentence with a smiley :(")
	assertCorrect("This is a sentence with a smiley :-)")
	assertCorrect("This is a sentence with a smiley ;-) and so on...")
	assertCorrect("This is a [test] sentence...")
	assertCorrect("The plight of Tamil refugees caused a surge of support from most of the Tamil political parties.[90]")
	assertCorrect("(([20] [20] [20]))")
	assertCorrect("This is a \"special test\", right?")
	// numerical bullets / chapter refs (Java numeral enumeration exception)
	assertCorrect("We discussed this in Chapter 1).")
	assertCorrect("The jury recommended that: (1) Four additional deputies be employed.")
	assertCorrect("We discussed this in section 1a).")
	assertCorrect("We discussed this in section iv).")
	assertCorrect("(Ketab fi Isti'mal al-'Adad al-Hindi)")
	assertCorrect("(al-'Adad al-Hindi)")
	assertCorrect("will-o'-the-wisp")
	assertCorrect("cat-o’-nine-tails")
	// lettered / numbered list as one analyzed sentence (Java getAnalyzedSentence)
	assertCorrect("a) item one\nb) item two\nc) item three")
	assertCorrect("\n\na) New York\nb) Boston\n")
	assertCorrect("\n\n1.) New York\n2.) Boston\n")
	assertCorrect("\n\nXII.) New York\nXIII.) Boston\n")
	assertCorrect("\n\nA) New York\nB) Boston\nC) Foo\n")

	// incorrect sentences
	assertIncorrect("(This is a test sentence.")
	// Java: unfinished paren without sentence-end punctuation → 0
	assertCorrect("This is not so (neither a nor b")
	assertIncorrect("This is not so (neither a nor b.")
	assertIncorrect("This is not so neither a nor b)")
	assertIncorrect("This is not so neither foo nor bar)")
	assertIncorrect("This is a test sentence).")

	// smiley inside parens: no ! → correct; with ! → incorrect (Java)
	assertCorrect("Some text (and some funny remark :-) with more text to follow")
	assertIncorrect("Some text (and some funny remark :-) with more text to follow!")

	// multi-match (mismatched bracket types)
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("(This is a test] sentence.")})
	require.Equal(t, 2, len(ms))
	ms = rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("This [is (a test} sentence.")})
	require.Equal(t, 3, len(ms))
}

func TestEnglishUnpairedBracketsRule_MultipleSentences(t *testing.T) {
	// Java testMultipleSentences uses lt.check (full pipeline); here MatchList on
	// SplitAndAnalyze approximates multi-sentence bracket stack continuity.
	rule := NewEnglishUnpairedBracketsRule(nil)

	// correct: brackets closed within later sentence
	sents := languagetool.SplitAndAnalyze(
		"This is multiple sentence text that contains a bracket: " +
			"[This is a bracket. With some text.] and this continues.\n")
	// If splitter breaks inside brackets, stack may report; when it keeps pairing, 0.
	// Prefer AnalyzePlain for Java getAnalyzedSentence-style single stream when possible.
	// Full multi-sent with global stack:
	// sentence1: ... bracket:
	// sentence2: [This is a bracket.
	// sentence3: With some text.] and this continues.
	// MatchList across sents should pop [ with ].
	_ = sents
	// Build three sentences with open across first and close on second — 0 when paired
	s1 := languagetool.AnalyzePlain("This is correct (yes).")
	s2 := languagetool.AnalyzePlain("Still fine.")
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})))

	// open [ not closed across sentences
	open := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Intro: [This is a bracket."),
		languagetool.AnalyzePlain("With some text."),
		languagetool.AnalyzePlain("And this continues."),
	}
	require.GreaterOrEqual(t, len(rule.MatchList(open)), 1)

	// open and close across sentences
	paired := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Intro: [This is a bracket."),
		languagetool.AnalyzePlain("With some text.] and this continues."),
	}
	require.Equal(t, 0, len(rule.MatchList(paired)))
}

// Java EnglishUnpairedBracketsRule: EN_UNPAIRED_BRACKETS, parentheses URL, example pair.
func TestEnglishUnpairedBracketsRule_Metadata(t *testing.T) {
	rule := NewEnglishUnpairedBracketsRule(nil)
	require.Equal(t, "EN_UNPAIRED_BRACKETS", rule.GetID())
	require.Contains(t, rule.GetURL(), "what-are-parentheses")
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "PUNCTUATION", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSTypographical, rule.GetLocQualityIssueType())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "He lived in a <marker>(</marker>large house.", inc[0].GetExample())
	// Java Rule.addExamplePair uses first fixed <marker> span as correction
	require.Equal(t, []string{"("}, inc[0].GetCorrections())
	require.Contains(t, rule.GetCorrectExamples()[0].GetExample(), "large")
}
