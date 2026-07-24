package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishUnpairedQuotesRuleTest.java
// Java uses tagged analysis; EnglishUnpairedQuotesRule POS-gates apostrophe exceptions
// (_apostrophe_contraction_ / POS / NNP). Tests inject those tags for contraction/inch
// surfaces; free-standing quotes stay untagged. Untagged AnalyzePlain is fail-closed.
import (
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// injectAposContractionTags marks ASCII/curly apostrophes that Java would treat as
// contractions (not quote marks) via _apostrophe_contraction_.
func injectAposContractionTags(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sent == nil {
		return nil
	}
	toks := sent.GetTokensWithoutWhitespace()
	for i, tok := range toks {
		if tok == nil {
			continue
		}
		t := tok.GetToken()
		if t != "'" && t != "’" {
			continue
		}
		prev := ""
		if i > 0 && toks[i-1] != nil {
			prev = toks[i-1].GetToken()
		}
		next := ""
		if i+1 < len(toks) && toks[i+1] != nil {
			next = toks[i+1].GetToken()
		}
		afterLetter := endsWithLetter(prev)
		// Arcminute/inch: ASCII ' after digit only (not curly ’ which pairs with ‘).
		afterDigit := t == "'" && endsWithDigit(prev)
		beforeLetter := startsWithLetter(next)
		beforeDigit := startsWithDigit(next)
		// 'till / 'em / '49: elision at word start
		elisionStart := (beforeLetter || beforeDigit) &&
			(i <= 1 || tok.IsWhitespaceBefore() || (toks[i-1] != nil && toks[i-1].IsSentenceStart()))
		// o'clock, d'Ivoire, al-'Adad: letter/hyphen-cluster before and letter after
		// with no space (mid-word apostrophe).
		midWord := !tok.IsWhitespaceBefore() && (afterLetter || endsWithHyphen(prev)) && beforeLetter
		if !afterLetter && !afterDigit && !elisionStart && !midWord {
			continue
		}
		pos := "_apostrophe_contraction_"
		tok.AddReading(languagetool.NewAnalyzedToken(t, &pos, nil), "twin")
	}
	return sent
}

func endsWithLetter(s string) bool {
	rs := []rune(s)
	return len(rs) > 0 && unicode.IsLetter(rs[len(rs)-1])
}

func startsWithLetter(s string) bool {
	rs := []rune(s)
	return len(rs) > 0 && unicode.IsLetter(rs[0])
}

func endsWithDigit(s string) bool {
	rs := []rune(s)
	return len(rs) > 0 && rs[len(rs)-1] >= '0' && rs[len(rs)-1] <= '9'
}

func startsWithDigit(s string) bool {
	rs := []rune(s)
	return len(rs) > 0 && rs[0] >= '0' && rs[0] <= '9'
}

func endsWithHyphen(s string) bool {
	rs := []rune(s)
	return len(rs) > 0 && rs[len(rs)-1] == '-'
}

func TestEnglishUnpairedQuotesRule_Rule(t *testing.T) {
	rule := NewEnglishUnpairedQuotesRule(nil)
	assertCorrect := func(s string) {
		t.Helper()
		sent := injectAposContractionTags(languagetool.AnalyzePlain(s))
		require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{sent})), "correct %q", s)
	}
	assertIncorrect := func(s string) {
		t.Helper()
		require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)})), "incorrect %q", s)
	}

	// correct sentences (Java EnglishUnpairedQuotesRuleTest)
	assertCorrect("This is a word 'test'.")
	assertCorrect("I don't know.")
	assertCorrect("This is the joint presidents' declaration.")
	assertCorrect("The screen is 20\"")
	assertCorrect("The screen is 20\" wide.")
	assertIncorrect("The screen is very\" wide.")
	assertCorrect("This is what he said: \"We believe in freedom. This is what we do.\"")

	assertCorrect("He was an ol' man.")
	assertCorrect("'till the end.")
	assertCorrect("jack-o'-lantern")
	assertCorrect("jack o'lantern")
	assertCorrect("sittin' there")
	assertCorrect("Nothin'")
	assertCorrect("ya'")
	assertCorrect("I'm not goin'")
	assertCorrect("y'know")
	assertCorrect("Please find attached Fritz' revisions")
	assertCorrect("You're only foolin' round.")
	assertCorrect("I stayed awake 'till the morning.")
	assertCorrect("under the 'Global Markets' heading")
	assertCorrect("He's an 'admin'.")
	assertCorrect("However, he's still expected to start in the 49ers' next game on Oct.")
	assertCorrect("all of his great-grandfathers' names")
	assertCorrect("Though EES' past profits now are in question")
	assertCorrect("Networks' Communicator and FocusFocus' Conference.")
	assertCorrect("Additional funding came from MegaMags' founders and existing individual investors.")
	assertCorrect("al-Jazā’er")
	assertCorrect("second Mu’taq and third")
	assertCorrect("second Mu'taq and third")
	assertCorrect("The phrase ‘1 2’ is British English.")

	assertCorrect("22' N., long. ")
	assertCorrect("11º 22'")
	assertCorrect("11° 22'")
	assertCorrect("11° 22.5'")
	assertCorrect("In case I garbled mine, here 'tis.")
	assertCorrect("It's about three o’clock.")
	assertCorrect("It's about three o'clock.")
	assertCorrect("Rory O’More")
	assertCorrect("Rory O'More")
	assertCorrect("Côte d’Ivoire")
	assertCorrect("Côte d'Ivoire")
	assertCorrect("Colonel d’Aubigni")
	assertCorrect("They are members of the Bahá'í Faith.")

	assertCorrect("This is a \"special test\", right?")
	assertCorrect("In addition, the government would pay a $1,000 \"cost of education\" grant to the schools.")
	assertCorrect("Paradise lost to the alleged water needs of Texas' big cities Thursday.")
	assertCorrect("Kill 'em all!")
	assertCorrect("Puttin' on the Ritz")
	assertCorrect("Dunkin' Donuts")
	assertCorrect("Hold 'em!")
	assertCorrect("(Ketab fi Isti'mal al-'Adad al-Hindi)")
	assertCorrect("(al-'Adad al-Hindi)")
	assertCorrect("On their 'host' societies.")
	assertCorrect("On their 'host society'.")
	assertCorrect("Burke-rostagno the Richard S. Burkes' home in Wayne may be the setting for the wedding reception for their daughter.")
	assertCorrect("The '49 team was off to a so-so 5-5 beginning")
	assertCorrect("The best reason that can be advanced for the state adopting the practice was the advent of expanded highway construction during the 1920s and '30s.")
	assertCorrect("A Republican survey says Kennedy won the '60 election on the religious issue.")
	assertCorrect("Economy class seats have a seat pitch of 31-33\", with newer aircraft having thinner seats that have a 31\" pitch.")
	assertCorrect("\"02\" will sort before \"10\" as expected so it will have size of 10\".")
	assertCorrect("\"02\" will sort before \"10\" as expected so it will have size of 10\"")
	assertCorrect("\"02\" will sort before \"10\"")
	assertCorrect("On their 'host societies'.")
	assertCorrect("On their host 'societies'.")
	assertIncorrect("On their 'host societies.")
	// Java TODO: ambiguous
	assertCorrect("On their host societies'.")
	assertCorrect("I think that Liszt's \"Forgotten Waltz No.3\" is a hidden masterpiece.")
	assertCorrect("I think that Liszt's \"Forgotten Waltz No. 3\" is a hidden masterpiece.")
	assertCorrect("Turkish distinguishes between dotted and dotless \"I\"s.")
	assertCorrect("It has recognized no \"bora\"-like pattern in his behaviour.")

	// incorrect
	assertIncorrect("This is a test with an apostrophe &'.")
	assertIncorrect("He is making them feel comfortable all along.\"")
	assertIncorrect("\"He is making them feel comfortable all along.")

	assertCorrect("Some text. This is \"12345\", a number.")
	assertCorrect("Some text. This is 12345\", a number.") // inch

	ms := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("\"This is a test” sentence.")})
	require.Equal(t, 2, len(ms))
	ms = rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("This \"is 'a test” sentence.")})
	require.Equal(t, 3, len(ms))
}

func TestEnglishUnpairedQuotesRule_FailClosedWithoutPOS(t *testing.T) {
	rule := NewEnglishUnpairedQuotesRule(nil)
	// Without POS inject, contraction apostrophes are treated as quotes (EN override).
	n := len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("\"I'm over here, she said.")}))
	require.Equal(t, 2, n)
}

func TestEnglishUnpairedQuotesRule_Metadata(t *testing.T) {
	rule := NewEnglishUnpairedQuotesRule(nil)
	require.Equal(t, "EN_UNPAIRED_QUOTES", rule.GetID())
	require.Contains(t, rule.GetURL(), "what-are-quotation-marks")
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "PUNCTUATION", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSTypographical, rule.GetLocQualityIssueType())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "\"I'm over here,<marker></marker> she said.", inc[0].GetExample())
	require.Equal(t, []string{"\""}, inc[0].GetCorrections())
	require.Equal(t, "\"I'm over here,<marker>\"</marker> she said.", rule.GetCorrectExamples()[0].GetExample())
}

func TestEnglishUnpairedQuotesRule_MultipleSentences(t *testing.T) {
	rule := NewEnglishUnpairedQuotesRule(nil)
	matchN := func(parts ...string) int {
		var as []*languagetool.AnalyzedSentence
		for _, p := range parts {
			as = append(as, languagetool.AnalyzePlain(p))
		}
		return len(rule.MatchList(as))
	}
	require.Equal(t, 0, matchN(
		"This is multiple sentence text that contains Quotes: \"This is a bracket.",
		"With some text.\" and this continues.\n",
	))
	require.Equal(t, 0, matchN(
		"This is multiple sentence text that contains Quotes. “This is a bracket.",
		"\n\n With some text.” and this continues.",
	))
	require.Equal(t, 1, matchN(
		"This is multiple sentence text that contains a Quote: “This is a bracket.",
		"With some text. And this continues.\n\n",
	))
}
