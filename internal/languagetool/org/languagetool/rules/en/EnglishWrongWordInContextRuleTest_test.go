package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishWrongWordInContextRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishWrongWordInContextRule_Rule(t *testing.T) {
	rule := NewEnglishWrongWordInContextRule(nil)
	assertGood := func(sentence string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(sentence))), "good: %q", sentence)
	}
	assertBad := func(sentence string) {
		t.Helper()
		require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain(sentence))), "bad: %q", sentence)
	}

	// prescribe/proscribe
	assertBad("I have proscribed you a course of antibiotics.")
	assertGood("I have prescribed you a course of antibiotics.")
	assertGood("Name one country that does not proscribe theft.")
	assertBad("Name one country that does not prescribe theft.")
	matches := rule.Match(languagetool.AnalyzePlain("I have proscribed you a course of antibiotics."))
	require.Equal(t, "prescribed", matches[0].GetSuggestedReplacements()[0])

	// herion/heroine
	assertBad("We know that heroine is highly addictive.")
	assertGood("He wrote about his addiction to heroin.")
	assertGood("A heroine is the principal female character in a novel.")
	assertBad("A heroin is the principal female character in a novel.")

	// bizarre/bazaar
	assertBad("What a bazaar behavior!")
	assertGood("I bought these books at the church bazaar.")
	assertGood("She has a bizarre haircut.")
	assertBad("The Saturday morning bizarre is worth seeing even if you buy nothing.")

	// bridal/bridle
	assertBad("The bridle party waited on the lawn.")
	assertGood("Forgo the champagne treatment a bridal boutique often provides.")
	assertGood("He sat there holding his horse by the bridle.")
	assertBad("Each rider used his own bridal.")

	// desert/dessert
	assertBad("They have some great deserts on this menu.")
	assertGood("They have some great desserts on this menu.")

	// statute/statue
	assertBad("They have some great marble statutes.")
	assertGood("They have a great marble statue.")

	// neutron/neuron
	assertGood("Protons and neutrons")
	assertBad("Protons and neurons")

	// hangar / hanger
	assertBad("The plane taxied to the hanger.")
	assertGood("The plane taxied to the hangar.")
}
