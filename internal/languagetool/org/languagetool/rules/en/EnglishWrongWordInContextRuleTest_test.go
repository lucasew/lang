package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishWrongWordInContextRuleTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
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
	// Java: assertEquals("prescribed", rule.match(...)[0].getSuggestedReplacements().get(0));
	matches := rule.Match(languagetool.AnalyzePlain("I have proscribed you a course of antibiotics."))
	require.Len(t, matches, 1)
	require.Equal(t, "prescribed", matches[0].GetSuggestedReplacements()[0])
	// Java-visible: span over "proscribed", SpecificIdRule id, long message with explanations.
	require.Equal(t, strings.Index("I have proscribed you a course of antibiotics.", "proscribed"), matches[0].GetFromPos())
	require.Equal(t, matches[0].GetFromPos()+len("proscribed"), matches[0].GetToPos())
	idRule, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "Java always wraps match in SpecificIdRule")
	require.Equal(t, tools.ToId("ENGLISH_WRONG_WORD_IN_CONTEXT_proscribed_prescribed", "en"), idRule.GetID())
	require.Equal(t, "commonly confused words: proscribed/prescribed", idRule.GetDescription())
	require.Contains(t, matches[0].GetMessage(), "prescribed")
	require.Contains(t, matches[0].GetMessage(), "proscribed")
	require.Contains(t, matches[0].GetMessage(), "to recommend or authorize")
	require.Contains(t, matches[0].GetMessage(), "to forbid, to limit or to banish")
	require.Equal(t, "Possibly confused word", matches[0].GetShortMessage())

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
	deserts := rule.Match(languagetool.AnalyzePlain("They have some great deserts on this menu."))
	require.Equal(t, "desserts", deserts[0].GetSuggestedReplacements()[0])

	// statute/statue
	assertBad("They have some great marble statutes.")
	assertGood("They have a great marble statue.")
	statutes := rule.Match(languagetool.AnalyzePlain("They have some great marble statutes."))
	require.Equal(t, "statues", statutes[0].GetSuggestedReplacements()[0])

	// neutron/neuron
	assertGood("Protons and neutrons")
	assertBad("Protons and neurons")
	neurons := rule.Match(languagetool.AnalyzePlain("Protons and neurons"))
	require.Equal(t, "neutrons", neurons[0].GetSuggestedReplacements()[0])

	// hangar / hanger
	// neutron/neuron (Java comment duplicated)
	assertBad("The plane taxied to the hanger.")
	assertGood("The plane taxied to the hangar.")
	hanger := rule.Match(languagetool.AnalyzePlain("The plane taxied to the hanger."))
	require.Equal(t, "hangar", hanger[0].GetSuggestedReplacements()[0])
}

// Java EnglishWrongWordInContextRule: metadata + proscribed → prescribed example pair.
func TestEnglishWrongWordInContextRule_Metadata(t *testing.T) {
	rule := NewEnglishWrongWordInContextRule(nil)
	require.Equal(t, "ENGLISH_WRONG_WORD_IN_CONTEXT", rule.GetID())
	require.Equal(t, "commonly confused words: $match", rule.GetDescription())
	require.Equal(t, "Commonly Confused Words", rule.GetCategory().GetName())
	require.Equal(t, rules.CategoryConfusedWords, rule.GetCategory().GetID())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "I have <marker>proscribed</marker> you a course of antibiotics.", inc[0].GetExample())
	require.Equal(t, []string{"prescribed"}, inc[0].GetCorrections())
	require.Equal(t, "I have <marker>prescribed</marker> you a course of antibiotics.", rule.GetCorrectExamples()[0].GetExample())
}
