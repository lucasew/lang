package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/DuUpperLowerCaseRuleTest.java
//
// Java:
//   DuUpperLowerCaseRule rule = new DuUpperLowerCaseRule(TestTools.getEnglishMessages());
//   JLanguageTool lt = new JLanguageTool(Languages.getLanguageForShortCode("de-DE"));
//   AnalyzedSentence sentence = lt.getAnalyzedSentence(input);
//   RuleMatch[] matches = rule.match(Collections.singletonList(sentence));
//
// Go: de-DE JLanguageTool.GetAnalyzedSentence (real DE analysis path) — no AnalyzePlain invent.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// newDuUpperLowerCaseFixture mirrors Java test field setup (rule + de-DE analysis).
func newDuUpperLowerCaseFixture(t *testing.T) (*DuUpperLowerCaseRule, *languagetool.JLanguageTool) {
	t.Helper()
	// Java: new DuUpperLowerCaseRule(TestTools.getEnglishMessages()) — category msgs only
	rule := NewDuUpperLowerCaseRule(nil)
	lt := languagetool.NewJLanguageTool("de-DE")
	// Real DE tagger path (Java createDefaultTagger); rule itself is surface-only but
	// analysis must match de-DE GetAnalyzedSentence token stream.
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for DuUpperLowerCaseRule twin")
	return rule, lt
}

// TestDuUpperLowerCaseRule_Rule ports DuUpperLowerCaseRuleTest.testRule full matrix.
func TestDuUpperLowerCaseRule_Rule(t *testing.T) {
	rule, lt := newDuUpperLowerCaseFixture(t)

	assertErrors := func(input string, expectedMatches int) {
		t.Helper()
		// Java: rule.match(Collections.singletonList(lt.getAnalyzedSentence(input)))
		sentence := lt.GetAnalyzedSentence(input)
		ms := rule.MatchList([]*languagetool.AnalyzedSentence{sentence})
		require.Equal(t, expectedMatches, len(ms), "input=%q matches=%v", input, ms)
	}

	// Java DuUpperLowerCaseRuleTest.testRule matrix (every assert has a twin):
	assertErrors("Du bist noch jung.", 0)
	assertErrors("Du bist noch jung, du bist noch fit.", 0)
	assertErrors("Aber du bist noch jung, du bist noch fit.", 0)
	assertErrors("Aber du bist noch jung, dir ist das egal.", 0)
	assertErrors("Hast Du ihre Brieftasche gesehen?", 0)

	assertErrors("Aber Du bist noch jung, du bist noch fit.", 1)
	assertErrors("Aber Du bist noch jung, dir ist das egal.", 1)
	assertErrors("Aber Du bist noch jung. Und dir ist das egal.", 1)

	assertErrors("Aber du bist noch jung. Und Du bist noch fit.", 1)
	assertErrors("Aber du bist noch jung, Dir ist das egal.", 1)
	assertErrors("Aber du bist noch jung. Und Dir ist das egal.", 1)

	assertErrors("Aber du bist noch jung, sagt euer Vater oft.", 0)
	assertErrors("Aber Du bist noch jung, sagt Euer Vater oft.", 0)
	assertErrors("Aber Du bist noch jung, sagt euer Vater oft.", 1)
	assertErrors("Aber du bist noch jung, sagt Euer Vater oft.", 1)

	assertErrors("Könnt Ihr Euch das vorstellen???", 0)
	assertErrors("Könnt ihr euch das vorstellen???", 0)
	assertErrors("Aber Samstags geht ihr Sohn zum Sport. Stellt Euch das mal vor!", 0)
	// commented out in Java:
	// assertErrors("Könnt Ihr euch das vorstellen???", 1)
	assertErrors("Wie geht es euch? Herr Meier, wie war ihr Urlaub?", 0)
	assertErrors("Wie geht es Euch? Herr Meier, wie war Ihr Urlaub?", 0)
	// commented out in Java:
	// assertErrors("Wie geht es euch? Herr Meier, wie war Ihr Urlaub?", 1)
	// assertErrors("Wie geht es Euch? Herr Meier, wie war ihr Urlaub?", 1)

	assertErrors("\"Du sagtest, du würdest es schaffen!\"", 0)
	assertErrors("Egal, was du tust: Du musst dein Bestes geben.", 0)
	assertErrors("Was auch immer du tust: ICH UND DU KÖNNEN ES SCHAFFEN.", 0)
	assertErrors("Hast Du die Adresse von ihr?", 0)
}

// Java-visible outcomes beyond match count: suggestion, span, message, meta.
func TestDuUpperLowerCaseRule_SuggestionAndMeta(t *testing.T) {
	rule, lt := newDuUpperLowerCaseFixture(t)

	// Class example (Dir then du) — Java constructor Example.wrong/fixed
	sentence := lt.GetAnalyzedSentence("Wie geht es Dir? Bist du wieder gesund?")
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{sentence})
	require.Equal(t, 1, len(ms))
	require.Equal(t, "DE_DU_UPPER_LOWER", ms[0].GetRule().(interface{ GetID() string }).GetID())
	require.Equal(t, []string{"Du"}, ms[0].GetSuggestedReplacements())
	require.Contains(t, ms[0].GetMessage(), "Dir")
	require.Contains(t, ms[0].GetMessage(), "großgeschrieben")
	require.Contains(t, ms[0].GetMessage(), "Du")
	// span on "du" token (Java pos+token.getStartPos/EndPos)
	var duTok *languagetool.AnalyzedTokenReadings
	for _, tok := range sentence.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "du" {
			duTok = tok
			break
		}
	}
	require.NotNil(t, duTok)
	require.Equal(t, duTok.GetStartPos(), ms[0].GetFromPos())
	require.Equal(t, duTok.GetEndPos(), ms[0].GetToPos())

	// Lower→upper flag: first "du", later "Euer"
	sentence2 := lt.GetAnalyzedSentence("Aber du bist noch jung, sagt Euer Vater oft.")
	ms2 := rule.MatchList([]*languagetool.AnalyzedSentence{sentence2})
	require.Equal(t, 1, len(ms2))
	require.Equal(t, []string{"euer"}, ms2[0].GetSuggestedReplacements())
	require.Contains(t, ms2[0].GetMessage(), "kleingeschrieben")
	require.Contains(t, ms2[0].GetMessage(), "euer")

	// constructor / TextLevelRule surface
	require.Equal(t, "DE_DU_UPPER_LOWER", rule.GetID())
	require.Equal(t, "Einheitliche Verwendung von Du/du, Dir/dir etc.", rule.GetDescription())
	require.Equal(t, -1, rule.MinToCheckParagraph())
	require.Equal(t, -1, rule.EstimateContextForSureMatch())
	require.Equal(t, "https://languagetool.org/insights/de/beitrag/duzen-grossgeschrieben/", rule.GetURL())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "CASING", rule.GetCategory().GetID().String())
	require.Len(t, rule.GetIncorrectExamples(), 1)
	require.Len(t, rule.GetCorrectExamples(), 1)
	require.Equal(t, []string{"Du"}, rule.GetIncorrectExamples()[0].GetCorrections())
	// compile-time: implements TextLevelRule
	var _ rules.TextLevelRule = rule
}

// Unit: Apache Commons StringUtils.isAllUpperCase used by Java match branch.
func TestDuUpperLowerCaseRule_StringUtilsIsAllUpperCase(t *testing.T) {
	require.True(t, stringUtilsIsAllUpperCase("DU"))
	require.True(t, stringUtilsIsAllUpperCase("ICH"))
	require.False(t, stringUtilsIsAllUpperCase("Du"))
	require.False(t, stringUtilsIsAllUpperCase("du"))
	require.False(t, stringUtilsIsAllUpperCase(""))
	require.False(t, stringUtilsIsAllUpperCase("DU!")) // Commons: every char must be upper
}
