package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/AdaptSuggestionFilterTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// adaptSuggestionFilterTestLT ports Java:
//   private final JLanguageTool lt = new JLanguageTool(Languages.getLanguageForShortCode("de"));
// with createDefaultTagger + createDefaultDisambiguator.
func adaptSuggestionFilterTestLT(t *testing.T) (*AdaptSuggestionFilter, *languagetool.JLanguageTool) {
	t.Helper()
	lt := languagetool.NewJLanguageTool("de-DE")
	if !WireGermanTagWord(lt) {
		t.Skip("German tagger resources unavailable")
	}
	WireGermanDisambiguator(lt)
	f := WireAdaptSuggestionFilter()
	require.NotNil(t, f.GenderOf, "GenderOf (GermanTagger)")
	require.NotNil(t, f.Synthesize, "Synthesize (GermanSynthesizer)")
	require.NotNil(t, f.FilterSuggestions, "FilterSuggestions (AgreementRule SuggestionFilter)")
	return f, lt
}

// Twin of AdaptSuggestionFilterTest.testAcceptRuleMatchDevTest (Java @Ignore for development).
func TestAdaptSuggestionFilter_AcceptRuleMatchDevTest(t *testing.T) {
	t.Skip("Java @Ignore(\"for development\")")
}

// Twin of AdaptSuggestionFilterTest.testAcceptRuleMatchWithDet.
func TestAdaptSuggestionFilter_AcceptRuleMatchWithDet(t *testing.T) {
	f, lt := adaptSuggestionFilterTestLT(t)
	// MAS (Der Plan):
	runAcceptRuleMatch(t, f, lt, "Die Roadmap ist gut.", "Roadmap", "Plan", "[Den Plan, Der Plan]")
	// MAS (der Plan):
	runAcceptRuleMatch(t, f, lt, "Hier steht die Roadmap.", "Roadmap", "Plan", "[den Plan, der Plan]")
	runAcceptRuleMatch(t, f, lt, "Hier steht eine Roadmap.", "Roadmap", "Plan", "[einen Plan, ein Plan]")
	runAcceptRuleMatch(t, f, lt, "Hier steht meine Roadmap.", "Roadmap", "Plan", "[meinen Plan, mein Plan]")
	runAcceptRuleMatch(t, f, lt, "Hier steht deine Roadmap.", "Roadmap", "Plan", "[deinen Plan, dein Plan]")
	runAcceptRuleMatch(t, f, lt, "Hier steht seine Roadmap.", "Roadmap", "Plan", "[seinen Plan, sein Plan]")
	runAcceptRuleMatch(t, f, lt, "Hier steht ihre Roadmap.", "Roadmap", "Plan", "[ihren Plan, ihr Plan]")
	runAcceptRuleMatch(t, f, lt, "Hier steht unsere Roadmap.", "Roadmap", "Plan", "[unseren Plan, unser Plan]")
	runAcceptRuleMatch(t, f, lt, "Hier steht eure Roadmap.", "Roadmap", "Plan", "[euren Plan, euer Plan]")
	// FEM (die Idee):
	runAcceptRuleMatch(t, f, lt, "Hier steht die Roadmap.", "Roadmap", "Idee", "[die Idee]")
	runAcceptRuleMatch(t, f, lt, "Hier steht eine Roadmap.", "Roadmap", "Idee", "[eine Idee]")
	runAcceptRuleMatch(t, f, lt, "Hier steht meine Roadmap.", "Roadmap", "Idee", "[meine Idee]")
	runAcceptRuleMatch(t, f, lt, "Hier steht deine Roadmap.", "Roadmap", "Idee", "[deine Idee]")
	runAcceptRuleMatch(t, f, lt, "Hier steht seine Roadmap.", "Roadmap", "Idee", "[seine Idee]")
	runAcceptRuleMatch(t, f, lt, "Hier steht ihre Roadmap.", "Roadmap", "Idee", "[ihre Idee]")
	runAcceptRuleMatch(t, f, lt, "Hier steht unsere Roadmap.", "Roadmap", "Idee", "[unsere Idee]")
	runAcceptRuleMatch(t, f, lt, "Hier steht eure Roadmap.", "Roadmap", "Idee", "[eure Idee]")
	// NEU (das Verfahren):
	runAcceptRuleMatch(t, f, lt, "Hier steht die Roadmap.", "Roadmap", "Verfahren", "[das Verfahren]")
	runAcceptRuleMatch(t, f, lt, "Hier steht eine Roadmap.", "Roadmap", "Verfahren", "[ein Verfahren]")
	runAcceptRuleMatch(t, f, lt, "Hier steht meine Roadmap.", "Roadmap", "Verfahren", "[mein Verfahren]")
	runAcceptRuleMatch(t, f, lt, "Hier steht deine Roadmap.", "Roadmap", "Verfahren", "[dein Verfahren]")
	runAcceptRuleMatch(t, f, lt, "Hier steht seine Roadmap.", "Roadmap", "Verfahren", "[sein Verfahren]")
	runAcceptRuleMatch(t, f, lt, "Hier steht ihre Roadmap.", "Roadmap", "Verfahren", "[ihr Verfahren]")
	runAcceptRuleMatch(t, f, lt, "Hier steht unsere Roadmap.", "Roadmap", "Verfahren", "[unser Verfahren]")
	runAcceptRuleMatch(t, f, lt, "Hier steht eure Roadmap.", "Roadmap", "Verfahren", "[euer Verfahren]")
}

// Twin of AdaptSuggestionFilterTest.testAcceptRuleMatchWithDetAdj (Java @Ignore WIP).
func TestAdaptSuggestionFilter_AcceptRuleMatchWithDetAdj(t *testing.T) {
	t.Skip("Java @Ignore(\"WIP\") — detAdjNoun path not active upstream (&& false)")
}

// Twin of AdaptSuggestionFilterTest.testAdaptedDet.
func TestAdaptSuggestionFilter_AdaptedDet(t *testing.T) {
	f, _ := adaptSuggestionFilterTestLT(t)
	assertDet(t, f, "die", "ART:DEF:NOM:SIN:FEM", "der", "Mann", "[der]")
	assertDet(t, f, "der", "ART:DEF:NOM:SIN:MAS", "der", "Frau", "[die]")
	assertDet(t, f, "der", "ART:DEF:NOM:SIN:NEU", "der", "Kind", "[das]")

	assertDet(t, f, "eine", "ART:IND:NOM:SIN:FEM", "ein", "Plan", "[ein]")     // eine Roadmap -> ein Plan
	assertDet(t, f, "eine", "ART:IND:AKK:SIN:FEM", "ein", "Plan", "[einen]")   // ich habe eine Roadmap -> ich habe einen Plan
	assertDet(t, f, "einer", "ART:IND:GEN:SIN:FEM", "ein", "Plan", "[eines]")  // die Ausführung einer Roadmap
	assertDet(t, f, "einer", "ART:IND:DAT:SIN:FEM", "ein", "Plan", "[einem]")  // einer Roadmap -> einem Plan
}

// Twin of AdaptSuggestionFilterTest.testdAdaptedDetAdj (Java method name has lowercase d; @Ignore WIP).
func TestAdaptSuggestionFilter_DAdaptedDetAdj(t *testing.T) {
	t.Skip("Java @Ignore(\"WIP\") — getAdaptedDetAdj not active in acceptRuleMatch")
}

// runAcceptRuleMatch ports AdaptSuggestionFilterTest.runAcceptRuleMatch.
func runAcceptRuleMatch(t *testing.T, filter *AdaptSuggestionFilter, lt *languagetool.JLanguageTool,
	sentenceStr, word, origReplacement, newReplacements string) {
	t.Helper()
	sentence := lt.GetAnalyzedSentence(sentenceStr)
	require.NotNil(t, sentence, "GetAnalyzedSentence(%q)", sentenceStr)
	fromPos := strings.Index(sentenceStr, word)
	require.GreaterOrEqual(t, fromPos, 0, "word %q in %q", word, sentenceStr)
	toPos := fromPos + len(word)
	match := rules.NewRuleMatch(rules.NewFakeRule(""), sentence, fromPos, toPos, "fake message")
	match.SetSuggestedReplacement(origReplacement)
	tokenPos := -1
	i := 0
	tokens := sentence.GetTokensWithoutWhitespace()
	for _, tok := range tokens {
		if tok != nil && tok.GetToken() == word {
			tokenPos = i
			break
		}
		i++
	}
	require.GreaterOrEqual(t, tokenPos, 0, "Word %q not found in sentence: %q", word, sentenceStr)
	// Java passes i after loop (equals tokenPos when found) and pattern tokens slice.
	args := map[string]string{"sub": `\1`}
	var patternToks []*languagetool.AnalyzedTokenReadings
	if tokenPos >= 0 && tokenPos < len(tokens) {
		patternToks = tokens[tokenPos : tokenPos+1]
	}
	newMatch := filter.AcceptRuleMatch(match, args, i, patternToks, nil)
	require.NotNil(t, newMatch)
	got := "[" + strings.Join(newMatch.GetSuggestedReplacements(), ", ") + "]"
	require.Equal(t, newReplacements, got, "sentence=%q word=%q repl=%q", sentenceStr, word, origReplacement)
}

// assertDet ports AdaptSuggestionFilterTest.assertDet.
func assertDet(t *testing.T, filter *AdaptSuggestionFilter, token, pos, lemma, replWord, expectedDet string) {
	t.Helper()
	adapted := filter.AdaptedDet(DetReading{Token: token, POS: pos, Lemma: lemma}, replWord)
	got := "[" + strings.Join(adapted, ", ") + "]"
	require.Equal(t, expectedDet, got, "det=%s pos=%s repl=%s", token, pos, replWord)
}
