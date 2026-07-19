package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishWordRepeatRuleTest.java
import (
	"strings"
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
	// POS-dependent: inject HasPartialPosTag targets (Java posIsIn); AnalyzePlain has no tagger.
	assertGoodTagged := func(s string, surfacePOS map[string]string) {
		t.Helper()
		sent := languagetool.AnalyzePlain(s)
		injectPOSBySurface(sent, surfacePOS)
		n := len(rule.Match(sent))
		require.Equal(t, 0, n, "assertGoodTagged %q got %d", s, n)
	}

	assertGood("This is a test.")
	assertGoodTagged("If I had had time, I would have gone to see him.", map[string]string{"I": "PRP"})
	assertGoodTagged("I don't think that that is a problem.", map[string]string{"is": "VBZ"})
	assertGoodTagged("He also said that Azerbaijan had fulfilled a task he set, which was that that their defense budget should exceed the entire state budget of Armenia.", map[string]string{"their": "PRP$"})
	assertGoodTagged("Just as if that was proof that that English was correct.", map[string]string{"English": "NN"})
	assertGoodTagged("It was noticed after more than a month that that promise had not been carried out.", map[string]string{"promise": "NN"})
	assertGoodTagged("It was said that that lady was an actress.", map[string]string{"lady": "NN"})
	assertGoodTagged("Kurosawa's three consecutive movies after Seven Samurai had not managed to capture Japanese audiences in the way that that film had.", map[string]string{"film": "NN"})
	assertGoodTagged("The can can hold the water.", map[string]string{"can": "NN"})
	assertGood("May May awake up?")
	assertGood("May may awake up.")
	assertGood("The cat does meow meow")
	assertGood("Hah Hah")
	assertGood("Hip Hip Hooray")
	assertBad("Hip Hip")
	assertGood("It's S.T.E.A.M.")
	assertGood("Ok ok ok!")
	assertGood("O O O")
	assertGoodTagged("Alice and Bob had had a long-standing relationship.", map[string]string{"Bob": "NN"})
	assertBad("I may may awake up.")
	assertBad("That is May May.")
	assertGood("Will Will awake up?")
	assertGood("Will will awake up.")
	assertBad("I will will awake up.")
	assertBad("Please wait wait for me.")
	assertGood("Wait wait!")
	assertBad("That is Will Will.")
	assertBad("I will will hold the ladder.")
	// follower "this" is DT — not in Java posIsIn list → error
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
	assertGoodTagged("I know that that can't really happen.", map[string]string{"can": "MD"})
	assertGoodTagged("Please pass her her phone.", map[string]string{"pass": "VB", "phone": "NN"})
	assertGood("I have visited Bora Bora.")
}

func TestEnglishWordRepeatRule_PosIsInFailClosed(t *testing.T) {
	// Without POS tags, "her her" / "had had" / "that that" / "can can" are not ignored (no invent).
	rule := NewEnglishWordRepeatRule(map[string]string{"repetition": "rep"})
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Please pass her her phone."))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("If I had had time."))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("The can can hold water."))))
}

// Java EnglishWordRepeatRule: ENGLISH_WORD_REPEAT_RULE id + is is example pair.
func TestEnglishWordRepeatRule_Metadata(t *testing.T) {
	rule := NewEnglishWordRepeatRule(nil)
	require.Equal(t, "ENGLISH_WORD_REPEAT_RULE", rule.GetID())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "This <marker>is is</marker> just an example sentence.", inc[0].GetExample())
	require.Equal(t, []string{"is"}, inc[0].GetCorrections())
	require.Equal(t, "This <marker>is</marker> just an example sentence.", rule.GetCorrectExamples()[0].GetExample())
}

// injectPOSBySurface adds POS readings to every non-whitespace token whose surface
// equals (case-insensitive) a map key — enough for Java posIsIn twin checks without a tagger.
// Special: "can" only tags the first occurrence (noun can before modal can).
func injectPOSBySurface(sent *languagetool.AnalyzedSentence, surfacePOS map[string]string) {
	nws := sent.GetTokensWithoutWhitespace()
	canFirstDone := false
	for i, tok := range nws {
		if tok == nil {
			continue
		}
		w := tok.GetToken()
		for surface, tag := range surfacePOS {
			if !strings.EqualFold(w, surface) {
				continue
			}
			if strings.EqualFold(surface, "can") {
				if canFirstDone {
					continue
				}
				canFirstDone = true
			}
			pos := tag
			nws[i].AddReading(languagetool.NewAnalyzedToken(w, &pos, nil), "test")
		}
	}
}
