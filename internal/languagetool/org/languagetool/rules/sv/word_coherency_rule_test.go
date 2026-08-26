package sv

// Behavior matrix for org.languagetool.rules.sv.WordCoherencyRule.
// No dedicated Java unit test in the SV module — outcomes derived from:
//   WordCoherencyRule.java (id/description/message/example pair)
//   AbstractWordCoherencyRule.match (text-level coherency)
//   official /org/languagetool/rules/sv/coherency.txt (mejl;mail, jos;juice, …)
//
// Surface forms match file pairs directly (AnalyzePlain). Lemmas are not required
// for the Java example pair; production loads file pairs only (no invent expand).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestWordCoherencyRule_Metadata(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	require.Equal(t, "SV_WORD_COHERENCY", rule.GetID())
	require.Equal(t, "Enhetlig och konsekvent stavning av ord när det finns stavningsvarianter att välja på.", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.NewCategoryId("MISC"), rule.GetCategory().GetID())
	require.Equal(t, -1, rule.MinToCheckParagraph())

	// Java addExamplePair(mejl/mail multi-marker wrong → fixed with first marker mejl)
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Det är en blandning av <marker>mejl</marker> och <marker>mail</marker> i det du skriver.", inc[0].GetExample())
	require.Equal(t, []string{"mejl"}, inc[0].GetCorrections())
	require.Equal(t,
		"Om du använder enbart <marker>mejl</marker> när du skriver <marker>mejl</marker> blir det mer konsekvent.",
		rule.GetCorrectExamples()[0].GetExample(),
	)
}

func TestWordCoherencyRule_DataFilePairs(t *testing.T) {
	// Official coherency.txt pairs (byte-identical resource under data/coherency.txt).
	d := loadCoherencyData()
	require.Contains(t, d.WordMap, "mejl")
	require.Contains(t, d.WordMap["mejl"], "mail")
	require.Contains(t, d.WordMap, "mail")
	require.Contains(t, d.WordMap["mail"], "mejl")
	require.Contains(t, d.WordMap, "jos")
	require.Contains(t, d.WordMap["jos"], "juice")
	require.Contains(t, d.WordMap, "co-location")
	require.Contains(t, d.WordMap["co-location"], "colocation")
	require.Contains(t, d.WordMap, "te")
	require.Contains(t, d.WordMap["te"], "the")
}

func TestWordCoherencyRule_Rule(t *testing.T) {
	rule := NewWordCoherencyRule(nil)

	assertGood := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
			languagetool.AnalyzePlain(s),
		})), "good: %q", s)
	}
	assertError := func(s string) {
		t.Helper()
		require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{
			languagetool.AnalyzePlain(s),
		})), "error: %q", s)
	}

	// consistent spelling (Java example pair basis)
	assertGood("Det är en blandning av mejl och mejl i det du skriver.")
	assertGood("Det är en blandning av mail och mail i det du skriver.")
	assertGood("Jag skickar mejl varje dag.")
	// mixed variants → 1 match (Java wrong example)
	assertError("Det är en blandning av mejl och mail i det du skriver.")
	assertError("Det är en blandning av mail och mejl i det du skriver.")
	// another official pair
	assertGood("Jag dricker jos. Hon dricker jos.")
	assertError("Jag dricker jos. Hon dricker juice.")
}

func TestWordCoherencyRule_MessageAndSuggestion(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	// Established mejl → later mail: message(word1=mail, word2=mejl), suggest mejl
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Det är en blandning av mejl och mail i det du skriver."),
	})
	require.Len(t, ms, 1)
	require.Equal(t, "Använd endast en av stavningsvarianterna 'mail' och 'mejl' i en och samma text.", ms[0].GetMessage())
	require.Equal(t, 32, ms[0].GetFromPos())
	require.Equal(t, 36, ms[0].GetToPos())
	require.Equal(t, []string{"mejl"}, ms[0].GetSuggestedReplacements())

	// reverse: established mail → later mejl
	ms = rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Det är en blandning av mail och mejl i det du skriver."),
	})
	require.Len(t, ms, 1)
	require.Equal(t, "Använd endast en av stavningsvarianterna 'mejl' och 'mail' i en och samma text.", ms[0].GetMessage())
	require.Equal(t, []string{"mail"}, ms[0].GetSuggestedReplacements())
}

func TestWordCoherencyRule_CallIndependence(t *testing.T) {
	// Separate match calls do not share shouldNotAppearWord state (Java AbstractWordCoherencyRule).
	assertGood := func(s string) {
		t.Helper()
		rule := NewWordCoherencyRule(nil)
		require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
			languagetool.AnalyzePlain(s),
		})))
	}
	assertGood("Jag skickar mejl.")
	assertGood("Han får mail.") // not noticed across independent calls
}

func TestWordCoherencyRule_RuleCompleteTexts(t *testing.T) {
	// Text-level MatchList over multiple analyzed sentences (document span).
	check := func(parts ...string) int {
		t.Helper()
		sents := make([]*languagetool.AnalyzedSentence, 0, len(parts))
		for _, p := range parts {
			sents = append(sents, languagetool.AnalyzePlain(p))
		}
		return len(NewWordCoherencyRule(nil).MatchList(sents))
	}
	require.Equal(t, 0, check("Jag skickar mejl.", " Han får mejl."))
	require.Equal(t, 1, check("Jag skickar mejl.", " Han får mail."))
	require.Equal(t, 1, check("Jag skickar mail.", " Han får mejl."))
	// single-sentence mixed
	require.Equal(t, 1, check("Det är en blandning av mejl och mail i det du skriver."))
}
