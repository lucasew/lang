package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/AvsAnRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAvsAnRule_Rule(t *testing.T) {
	rule := NewAvsAnRule(nil)
	assertCorrect := func(s string) {
		t.Helper()
		m := rule.Match(languagetool.AnalyzePlain(s))
		require.Equal(t, 0, len(m), "assertCorrect %q got %d matches", s, len(m))
	}
	assertIncorrect := func(s string) {
		t.Helper()
		m := rule.Match(languagetool.AnalyzePlain(s))
		require.Equal(t, 1, len(m), "assertIncorrect %q got %d matches", s, len(m))
	}

	assertCorrect("It must be an xml name.")
	assertCorrect("analyze an hprof file")
	assertCorrect("This is an sbt project.")
	assertCorrect("Import an Xcode project.")
	assertCorrect("This is a oncer.")
	assertCorrect("She was a Oaxacan chef.")
	assertCorrect("The doctor requested a urinalysis.")
	assertCorrect("She brought a Ouija board.")
	assertCorrect("This is a test sentence.")
	assertCorrect("It was an hour ago.")
	assertCorrect("A university is ...")
	assertCorrect("A one-way street ...")
	assertCorrect("An hour's work ...")
	assertCorrect("Going to an \"industry party\".")
	assertCorrect("An 8-year old boy ...")
	assertCorrect("An 18-year old boy ...")
	assertCorrect("The A-levels are ...")
	assertCorrect("An NOP check ...")
	assertCorrect("A USA-wide license ...")
	assertCorrect("...asked a UN member.")
	assertCorrect("In an un-united Germany...")
	assertCorrect("Here, a and b are supplementary angles.")
	assertCorrect("The Qur'an was translated into Polish.")
	assertCorrect("See an:Grammatica")
	assertCorrect("See http://www.an.com")
	assertCorrect("Station A equals station B.")
	assertCorrect("e.g., the case endings -a -i -u and mood endings -u -a")
	assertCorrect("A'ight, y'all.")
	assertCorrect("He also wrote the comic strips Abbie an' Slats.")
	assertCorrect("Do an ngram tokenization fix.")
	assertCorrect("A car with a unibody construction.")
	assertCorrect("Given a Userset, create and return a matching PrincipalModel.")
	assertCorrect("An uninterpreted dream is like an unopened letter.")
	assertCorrect("He has been a utilities lawyer for over 30 years.")
	assertCorrect("He has a unibrow.")

	assertIncorrect("It was a hour ago.")
	assertIncorrect("It was an sentence that's long.")
	assertIncorrect("It was a uninteresting talk.")
	assertIncorrect("An university")
	assertIncorrect("A unintersting ...")
	assertIncorrect("A hour's work ...")
	assertIncorrect("Going to a \"industry party\".")
	assertIncorrect("It was a unidentifiable object.")
	assertIncorrect(" I'll set you up with an userid and password so you can pick the...")
	matches := rule.Match(languagetool.AnalyzePlain("It was a uninteresting talk with an long sentence."))
	require.Equal(t, 2, len(matches))

	assertCorrect("A University")
	assertCorrect("A Europe wide something")

	assertIncorrect("then an University sdoj fixme sdoopsd")
	assertIncorrect("A 8-year old boy ...")
	assertIncorrect("A 18-year old boy ...")
	assertIncorrect("...asked an UN member.")
	assertIncorrect("In a un-united Germany...")

	assertCorrect("A. R.J. Turgot")
	assertCorrect("Make sure that 3.a as well as 3.b are correct.")
	assertCorrect("Anyone for an MSc?")
	assertIncorrect("Anyone for a MSc?")
	assertCorrect("Anyone for an XMR-based writer?")
	assertCorrect("Its name in English is a[1] (), plural A's, As, as, or a's.")
	assertCorrect("An historic event")
	assertCorrect("A historic event")
}

func TestAvsAnRule_Suggestions(t *testing.T) {
	rule := NewAvsAnRule(nil)
	require.Equal(t, "a string", rule.SuggestAorAn("string"))
	require.Equal(t, "a university", rule.SuggestAorAn("university"))
	require.Equal(t, "an hour", rule.SuggestAorAn("hour"))
	require.Equal(t, "an all-terrain", rule.SuggestAorAn("all-terrain"))
	require.Equal(t, "a UNESCO", rule.SuggestAorAn("UNESCO"))
	require.Equal(t, "a historical", rule.SuggestAorAn("historical"))
}

func TestAvsAnRule_GetCorrectDeterminerFor(t *testing.T) {
	getDeterminerFor := func(word string) Determiner {
		token := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(word, nil, nil))
		return GetCorrectDeterminerFor(token)
	}
	require.Equal(t, DeterminerA, getDeterminerFor("string"))
	require.Equal(t, DeterminerA, getDeterminerFor("university"))
	require.Equal(t, DeterminerA, getDeterminerFor("UNESCO"))
	require.Equal(t, DeterminerA, getDeterminerFor("one-way"))
	require.Equal(t, DeterminerAN, getDeterminerFor("interesting"))
	require.Equal(t, DeterminerAN, getDeterminerFor("hour"))
	require.Equal(t, DeterminerAN, getDeterminerFor("all-terrain"))
	require.Equal(t, DeterminerAOrAN, getDeterminerFor("historical"))
	require.Equal(t, DeterminerUnknown, getDeterminerFor(""))
	require.Equal(t, DeterminerUnknown, getDeterminerFor("-way"))
	require.Equal(t, DeterminerUnknown, getDeterminerFor("camelCase"))
}

func TestAvsAnRule_GetCorrectDeterminerForException(t *testing.T) {
	require.Panics(t, func() {
		GetCorrectDeterminerFor(nil)
	})
}

func TestAvsAnRule_Positions(t *testing.T) {
	rule := NewAvsAnRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("a industry standard."))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 1, matches[0].GetToPos())

	matches = rule.Match(languagetool.AnalyzePlain("a \"industry standard\"."))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 1, matches[0].GetToPos())

	matches = rule.Match(languagetool.AnalyzePlain("a “industry standard”."))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 1, matches[0].GetToPos())

	matches = rule.Match(languagetool.AnalyzePlain("a ‘industry standard’."))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 1, matches[0].GetToPos())

	matches = rule.Match(languagetool.AnalyzePlain("a - industry standard\"."))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 1, matches[0].GetToPos())

	matches = rule.Match(languagetool.AnalyzePlain("This is a \"industry standard\"."))
	require.Equal(t, 8, matches[0].GetFromPos())
	require.Equal(t, 9, matches[0].GetToPos())

	matches = rule.Match(languagetool.AnalyzePlain("\"a industry standard\"."))
	require.Equal(t, 1, matches[0].GetFromPos())
	require.Equal(t, 2, matches[0].GetToPos())

	matches = rule.Match(languagetool.AnalyzePlain("\"Many say this is a industry standard\"."))
	require.Equal(t, 18, matches[0].GetFromPos())
	require.Equal(t, 19, matches[0].GetToPos())

	matches = rule.Match(languagetool.AnalyzePlain("Like many \"an desperado\" before him, Bart headed south into Mexico."))
	require.Equal(t, 11, matches[0].GetFromPos())
	require.Equal(t, 13, matches[0].GetToPos())
}
