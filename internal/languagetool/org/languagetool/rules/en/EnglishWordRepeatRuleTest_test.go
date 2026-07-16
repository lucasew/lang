package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishWordRepeatRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishWordRepeatRule_RepeatRule(t *testing.T) {
	rule := NewEnglishWordRepeatRule(map[string]string{
		"repetition":            "Word repetition",
		"desc_repetition":       "Word repetition",
		"desc_repetition_short": "repetition",
	})
	assertGood := func(s string) {
		t.Helper()
		n := len(rule.Match(languagetool.AnalyzePlain(s)))
		require.Equal(t, 0, n, "assertGood %q got %d", s, n)
	}
	assertBad := func(s string) {
		t.Helper()
		n := len(rule.Match(languagetool.AnalyzePlain(s)))
		require.Equal(t, 1, n, "assertBad %q got %d", s, n)
	}

	assertGood("This is a test.")
	assertGood("If I had had time, I would have gone to see him.")
	assertGood("I don't think that that is a problem.")
	assertGood("He also said that Azerbaijan had fulfilled a task he set, which was that that their defense budget should exceed the entire state budget of Armenia.")
	assertGood("Just as if that was proof that that English was correct.")
	assertGood("It was noticed after more than a month that that promise had not been carried out.")
	assertGood("It was said that that lady was an actress.")
	assertGood("Kurosawa's three consecutive movies after Seven Samurai had not managed to capture Japanese audiences in the way that that film had.")
	assertGood("The can can hold the water.")
	assertGood("May May awake up?")
	assertGood("May may awake up.")
	assertGood("The cat does meow meow")
	assertGood("Hah Hah")
	assertGood("Hip Hip Hooray")
	assertBad("Hip Hip")
	assertGood("It's S.T.E.A.M.")
	assertGood("Ok ok ok!")
	assertGood("O O O")
	assertGood("Alice and Bob had had a long-standing relationship.")
	assertBad("I may may awake up.")
	assertBad("That is May May.")
	assertGood("Will Will awake up?")
	assertGood("Will will awake up.")
	assertBad("I will will awake up.")
	assertBad("Please wait wait for me.")
	assertGood("Wait wait!")
	assertBad("That is Will Will.")
	assertBad("I will will hold the ladder.")
	assertBad("You can feel confident that that this administration will continue to support a free and open Internet.")
	assertBad("This is is a test.")
	assertGood("b a s i c a l l y")
	assertGood("You can contact E.ON on Instagram.")
	assertBad("But I i was not sure.")
	assertBad("I I am the best.")
	assertGood("In a land far far away.")
	assertGood("I love you so so much.")
	assertGood("Aye aye, sir!")
	assertGood("What Tom did didn't seem to bother Mary at all.")
	assertGood("Whatever you do don't leave the lid up on the toilet!")
	assertGood("Keep your chin up and whatever you do don't doubt yourself or your actions.")
	assertGood("I know that that can't really happen.")
	assertGood("Please pass her her phone.")
	assertGood("I have visited Bora Bora.")
}
