package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/AvsAnRuleTest.java
// Inject DT on a/an like English tagger (Java hasPosTag("DT") only — no surface invent).
import (
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestAvsAnRule_Rule(t *testing.T) {
	rule := NewAvsAnRule(nil)
	assertCorrect := func(s string) {
		t.Helper()
		m := rule.Match(analyzeAvsAn(s))
		require.Equal(t, 0, len(m), "assertCorrect %q got %d matches", s, len(m))
	}
	assertIncorrect := func(s string) {
		t.Helper()
		m := rule.Match(analyzeAvsAn(s))
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
	matches := rule.Match(analyzeAvsAn("It was a uninteresting talk with an long sentence."))
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
	require.Equal(t, DeterminerA, getDeterminerFor("string"))
	require.Equal(t, DeterminerA, getDeterminerFor("university"))
	require.Equal(t, DeterminerAN, getDeterminerFor("hour"))
	require.Equal(t, DeterminerAN, getDeterminerFor("all-terrain"))
	require.Equal(t, DeterminerA, getDeterminerFor("UNESCO"))
	require.Equal(t, DeterminerAOrAN, getDeterminerFor("historical"))
	require.Equal(t, DeterminerUnknown, getDeterminerFor(""))
	require.Equal(t, DeterminerUnknown, getDeterminerFor("-"))
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
	matches := rule.Match(analyzeAvsAn("a industry standard."))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 1, matches[0].GetToPos())

	matches = rule.Match(analyzeAvsAn("a \"industry standard\"."))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 1, matches[0].GetToPos())

	matches = rule.Match(analyzeAvsAn("a “industry standard”."))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 1, matches[0].GetToPos())

	matches = rule.Match(analyzeAvsAn("a ‘industry standard’."))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 1, matches[0].GetToPos())

	matches = rule.Match(analyzeAvsAn("a - industry standard\"."))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 1, matches[0].GetToPos())

	matches = rule.Match(analyzeAvsAn("This is a \"industry standard\"."))
	require.Equal(t, 8, matches[0].GetFromPos())
	require.Equal(t, 9, matches[0].GetToPos())

	matches = rule.Match(analyzeAvsAn("\"a industry standard\"."))
	require.Equal(t, 1, matches[0].GetFromPos())
	require.Equal(t, 2, matches[0].GetToPos())

	matches = rule.Match(analyzeAvsAn("\"Many say this is a industry standard\"."))
	require.Equal(t, 18, matches[0].GetFromPos())
	require.Equal(t, 19, matches[0].GetToPos())

	matches = rule.Match(analyzeAvsAn("Like many \"an desperado\" before him, Bart headed south into Mexico."))
	require.Equal(t, 11, matches[0].GetFromPos())
	require.Equal(t, 13, matches[0].GetToPos())
}

// Java AvsAnRule ctor: setUrl, Categories.MISC, Misspelling, example pair, description.
func TestAvsAnRule_Metadata(t *testing.T) {
	rule := NewAvsAnRule(nil)
	require.Equal(t, "Use of 'a' vs. 'an'", rule.GetDescription())
	require.Contains(t, rule.GetURL(), "indefinite-articles")
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "MISC", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())
	require.Equal(t, 1, rule.EstimateContextForSureMatch())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "The train arrived <marker>a hour</marker> ago.", inc[0].GetExample())
	require.Equal(t, []string{"an hour"}, inc[0].GetCorrections())
	require.Equal(t, "The train arrived <marker>an hour</marker> ago.", rule.GetCorrectExamples()[0].GetExample())
}

func getDeterminerFor(w string) Determiner {
	return GetCorrectDeterminerFor(languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(w, nil, nil)))
}

// analyzeAvsAn injects DT on indefinite articles like the English tagger.
func analyzeAvsAn(text string) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzePlain(text)
	tokens := sent.GetTokensWithoutWhitespace()
	for i, tok := range tokens {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		low := strings.ToLower(tok.GetToken())
		if low != "a" && low != "an" {
			continue
		}
		// Non-article without space: Qur'an, 3.a — allow sentence-start / open quotes
		if !tok.IsWhitespaceBefore() && i > 0 {
			prev := tokens[i-1]
			prevTok := prev.GetToken()
			afterSentStart := prev.IsSentenceStart()
			afterOpenQuote := prevTok == "\"" || prevTok == "\u201c" || prevTok == "\u2018" ||
				prevTok == "(" || prevTok == "["
			if !afterSentStart && !afterOpenQuote {
				continue
			}
		}
		if i+1 < len(tokens) {
			n := tokens[i+1].GetToken()
			if (n == "'" || n == "\u2019") && !tokens[i+1].IsWhitespaceBefore() {
				continue // an'
			}
		}
		// "a and b" letter variables
		if low == "a" && i+2 < len(tokens) {
			mid := strings.ToLower(tokens[i+1].GetToken())
			nxt := tokens[i+2].GetToken()
			if (mid == "and" || mid == "or" || mid == "equals" || mid == "equal") &&
				len([]rune(nxt)) == 1 && unicode.IsLetter([]rune(nxt)[0]) {
				continue
			}
		}
		pos := "DT"
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, nil), "test")
	}
	return sent
}
