package ar

// Twin of org.languagetool.rules.ar.ArabicWrongWordInContextRule.
// No upstream ArabicWrongWordInContextRuleTest.java (file comment: "do not forget to add tests");
// twin from Java-visible metadata + empty official wrongWordInContext.txt Match behavior + example pair.
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START); MatchLemmas=false on base.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java ArabicWrongWordInContextRule: id, description, category, messages, filename, example.
func TestArabicWrongWordInContextRule_Metadata(t *testing.T) {
	rule := NewArabicWrongWordInContextRule(nil)
	require.Equal(t, ArabicWrongWordInContextRuleID, rule.GetID())
	require.Equal(t, "ARABIC_WRONG_WORD_IN_CONTEXT", rule.GetID())
	// Exact Java string — no space after شائعة.
	require.Equal(t, "كلمات شائعةمتشابهة (ظل/ضلَ, رؤيا/رؤية الخ.)", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.CategoryConfusedWords, rule.GetCategory().GetID())
	require.Equal(t, "كلمات متشابهة", rule.GetCategory().GetName())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())
	require.Equal(t, "ar", rule.LanguageCode)
	require.Equal(t, ArabicWrongWordInContextFilePath, "/ar/wrongWordInContext.txt")
	require.Equal(t, "احتمال كلمة متشابهة: هل تقصد <suggestion>$SUGGESTION</suggestion> بدلا من '$WRONGWORD'?", rule.MessageString)
	require.Equal(t, "احتمال كلمة متشابهة", rule.ShortMessageString)
	require.Equal(t, "احتمال كلمة متشابهة: هل تقصد <suggestion>$SUGGESTION</suggestion> (= $EXPLANATION_SUGGESTION) بدلا من '$WRONGWORD' (= $EXPLANATION_WRONGWORD)?", rule.LongMessageString)

	// Official /ar/wrongWordInContext.txt is header comments only → zero loaded pairs.
	// Do not invent الضن/الظن data rows even though the example pair mentions them.
	require.Empty(t, rule.Entries)
	require.Len(t, rule.Entries, 0)

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "من سوء <marker>الضن</marker> بالله ترك الأمر بالمعروف.", inc[0].GetExample())
	require.Equal(t, []string{"الظن"}, inc[0].GetCorrections())
	cor := rule.GetCorrectExamples()
	require.Len(t, cor, 1)
	require.Equal(t, "من سوء <marker>الظن</marker> بالله ترك الأمر بالمعروف.", cor[0].GetExample())
}

// Match returns nothing: official data has no context pairs (header-only resource).
func TestArabicWrongWordInContextRule_EmptyDataMatch(t *testing.T) {
	rule := NewArabicWrongWordInContextRule(nil)
	require.Empty(t, rule.Entries)

	// Java example sentence (wrong marker stripped) — still 0 matches without data rows.
	exampleWrong := "من سوء الضن بالله ترك الأمر بالمعروف."
	exampleFixed := "من سوء الظن بالله ترك الأمر بالمعروف."
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(exampleWrong))), "example wrong text")
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(exampleFixed))), "example fixed text")

	// Arbitrary Arabic / Latin text — no pairs → no matches.
	for _, s := range []string{
		"ظل",
		"ضل",
		"رؤيا رؤية",
		"hello world",
		"",
		"من سوء الضن بالله ترك الأمر بالمعروف. من سوء الظن بالله.",
	} {
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(s))), "text: %q", s)
	}
}
