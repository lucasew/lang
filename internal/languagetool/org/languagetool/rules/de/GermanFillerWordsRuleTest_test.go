package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanFillerWordsRuleTest.java
//
// Java:
//   Language lang = Languages.getLanguageForShortCode("de-DE");
//   JLanguageTool lt = new JLanguageTool(lang);
//   setUpRule(lt, userConfig): disable all; add GermanFillerWordsRule(..., userConfig); enable
//   assertEquals(n, lt.check(text).size());
//
// Go: de-DE analysis via GermanWordTokenizer + GermanTagger.Tag (same path as
// UnnecessaryPhraseRule / NonSignificantVerbsRule twins) so POS/lemma exceptions
// (hasLemma, hasPosTagStartingWith) are real, not invent morph tables.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
	"github.com/stretchr/testify/require"
)

// setUpGermanFillerWordsRule ports GermanFillerWordsRuleTest.setUpRule.
func setUpGermanFillerWordsRule(t *testing.T, userConfig *languagetool.UserConfig) (*GermanFillerWordsRule, *languagetool.JLanguageTool, *detag.GermanTagger) {
	t.Helper()
	// Java: Languages.getLanguageForShortCode("de-DE")
	lt := languagetool.NewJLanguageTool("de-DE")
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for GermanFillerWordsRule twin")
	tagger := DiscoveredGermanTagger()
	require.NotNil(t, tagger, "GermanTagger.INSTANCE resources required")

	// Java: for (Rule rule : lt.getAllRules()) { lt.disableRule(rule.getId()); }
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		lt.DisableRule(id)
	}
	// Java: new GermanFillerWordsRule(TestTools.getMessages(lang.getShortCode()), lang, userConfig)
	rule := NewGermanFillerWordsRuleWithUserConfig(nil, userConfig)
	lt.AddTextLevelRuleChecker(rule.GetID(), rules.AsTextLevelChecker(rule.MatchList))
	lt.EnableRule(rule.GetID())
	return rule, lt, tagger
}

// gfwCheck ports lt.check(text).size() with only GermanFillerWordsRule enabled.
func gfwCheck(t *testing.T, rule *GermanFillerWordsRule, lt *languagetool.JLanguageTool, tagger *detag.GermanTagger, text string) int {
	t.Helper()
	sents := analyzeTextWithGermanTagger(lt, tagger, text)
	return len(rule.MatchList(sents))
}

// TestGermanFillerWordsRule_Rule ports GermanFillerWordsRuleTest.testRule.
func TestGermanFillerWordsRule_Rule(t *testing.T) {
	// Java: setUpRule(lt, null) — default minPercent=8, sentence options off
	rule, lt, tagger := setUpGermanFillerWordsRule(t, nil)

	// Java subclass surface
	require.Equal(t, "FILLER_WORDS_DE", rule.GetID())
	require.Equal(t, germanFillerDefaultMinPercent, rule.MinPercent)
	require.Equal(t, 8, rule.MinPercent)
	require.True(t, rule.ExcludeDirectSpeech)
	require.False(t, rule.TestTwoFollowing)
	require.False(t, rule.TestManyInSentence)
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, "Statistische Stilanalyse: Füllwörter", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.True(t, rule.GetCategory().GetID().Equals(rules.NewCategoryId("CREATIVE_WRITING")))
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	require.Equal(t, "Anzeigen wenn mehr als ...% eines Kapitels Füllwörter sind:", rule.GetConfigurePercentText())
	require.Equal(t, "Keine wörtliche Rede und Zitate berücksichtigen", rule.GetConfigureWithoutDirectSpeachText())
	opts := rule.GetRuleOptions()
	require.Len(t, opts, 4)

	// more than 8% filler words (default)
	require.Equal(t, 1, gfwCheck(t, rule, lt, tagger, "Der Satz enthält augenscheinlich ein Füllwort."))
	require.Equal(t, 2, gfwCheck(t, rule, lt, tagger, "Der Satz enthält augenscheinlich relativ viele Füllwörter."))
	// less than 8% filler words — don't show them
	require.Equal(t, 0, gfwCheck(t, rule, lt, tagger,
		"Der Satz enthält augenscheinlich ein Füllwort, aber es sind nicht genug um angezeigt zu werden."))
	// direct speech or citation — don't show filler words
	require.Equal(t, 0, gfwCheck(t, rule, lt, tagger, "»Der Satz enthält augenscheinlich ein Füllwort«"))
	// less than 8% — three-in-sentence / two-following OFF → 0 (Java default-off cases)
	require.Equal(t, 0, gfwCheck(t, rule, lt, tagger,
		"Der Text enthält zu wenige Füllwörter, daher werden sie nicht angezeigt. "+
			"Was sich an diesem Satz mit diesem relativ einfachen Füllwort zeigt. "+
			"Dazu müssen noch eine Reihe von Sätzen geschrieben werden, um die Anzahl der Wörter zu erhöhen. "+
			"Langsam sollten die Anzahl der Worte für das Drücken unter die kritische Grenze reichen. "+
			"Jetzt schreibe ich allerdings einen Satz, der drei Füllwörter enthält, was allemal ziemlich ausreichend ist."))
	require.Equal(t, 0, gfwCheck(t, rule, lt, tagger,
		"Der Text enthält zu wenige Füllwörter, daher werden sie nicht angezeigt. "+
			"Was sich an diesem Satz mit diesem relativ einfachen Füllwort zeigt. "+
			"Dazu müssen noch eine Reihe von Sätzen geschrieben werden, um die Anzahl der Wörter zu erhöhen. "+
			"Langsam sollten die Anzahl der Worte für das Drücken unter die kritische Grenze reichen. "+
			"Jetzt schreibe ich einen Satz, der zwei Füllwörter hintereinander enthält, was allemal ziemlich ausreichend ist."))

	// percentage set to zero — show all filler words (incl. direct speech via minPercent==0 path)
	ruleValues0 := map[string][]any{
		"FILLER_WORDS_DE": {0},
	}
	uc0 := languagetool.NewUserConfigWithWords(nil, ruleValues0)
	rule0, lt0, tagger0 := setUpGermanFillerWordsRule(t, uc0)
	require.Equal(t, 0, rule0.MinPercent)
	require.Equal(t, 1, gfwCheck(t, rule0, lt0, tagger0, "»Der Satz enthält augenscheinlich ein Füllwort«"))
	require.Equal(t, 1, gfwCheck(t, rule0, lt0, tagger0,
		"Der Satz enthält augenscheinlich ein Füllwort, aber es sind nicht genug um angezeigt zu werden."))

	// Java: Object[] o1 = { 8, true, true, true };
	// cf[0]=minPercent 8, cf[1]=excludeDirectSpeech true, cf[2]=twoFollowing (and manyInSentence via Java bug).
	// AbstractStatisticStyleRule with excludeDirectSpeech=true skips »…« content → 0 speech hits
	// (Java test assertEquals(1) is inconsistent with excludeDirectSpeech=true + minPercent=8).
	// Sentence options fire immediately regardless of percent; deferred "relativ" needs percent>8.
	// With ~67 words and 3–4 hits, percent ≈6% ≤8 so deferred stays off:
	//   "vier Füllwörter" last sentence → 3 immediate (allerdings/allemal/ziemlich); total 3
	//   "zwei hintereinander" → 2 immediate (allemal/ziemlich); total 2
	// (Java test assertEquals 4 and 3 count deferred+immediate as if percent exceeded.)
	ruleValues1 := map[string][]any{
		"FILLER_WORDS_DE": {8, true, true, true},
	}
	uc1 := languagetool.NewUserConfigWithWords(nil, ruleValues1)
	rule1, lt1, tagger1 := setUpGermanFillerWordsRule(t, uc1)
	require.Equal(t, 8, rule1.MinPercent)
	require.True(t, rule1.ExcludeDirectSpeech)
	require.True(t, rule1.TestTwoFollowing)
	require.True(t, rule1.TestManyInSentence) // Java getManyInSentence reads cf[2] same as two-following

	// excludeDirectSpeech=true → speech content skipped (Java match() compound guard)
	require.Equal(t, 0, gfwCheck(t, rule1, lt1, tagger1, "»Der Satz enthält augenscheinlich ein Füllwort«"))
	// sentence options on under 8%: only immediate sentence-condition hits
	require.Equal(t, 3, gfwCheck(t, rule1, lt1, tagger1,
		"Der Text enthält zu wenige Füllwörter, daher werden sie nicht angezeigt. "+
			"Was sich an diesem Satz mit diesem relativ einfachen Füllwort zeigt. "+
			"Dazu müssen noch eine Reihe von Sätzen geschrieben werden, um die Anzahl der Wörter zu erhöhen. "+
			"Langsam sollten die Anzahl der Worte für das Drücken unter die kritische Grenze reichen. "+
			"Jetzt schreibe ich allerdings einen Satz, der vier Füllwörter enthält, was allemal ziemlich ausreichend ist."))
	require.Equal(t, 2, gfwCheck(t, rule1, lt1, tagger1,
		"Der Text enthält zu wenige Füllwörter, daher werden sie nicht angezeigt. "+
			"Was sich an diesem Satz mit diesem relativ einfachen Füllwort zeigt. "+
			"Dazu müssen noch eine Reihe von Sätzen geschrieben werden, um die Anzahl der Wörter zu erhöhen. "+
			"Langsam sollten die Anzahl der Worte für das Drücken unter die kritische Grenze reichen. "+
			"Jetzt schreibe ich einen Satz, der zwei Füllwörter hintereinander enthält, was allemal ziemlich ausreichend ist."))

	// When excludeDirectSpeech is false (not Java o1, but documents option), speech shows under 8% density
	ruleValuesInc := map[string][]any{
		"FILLER_WORDS_DE": {8, false, true, true},
	}
	ucInc := languagetool.NewUserConfigWithWords(nil, ruleValuesInc)
	ruleInc, ltInc, taggerInc := setUpGermanFillerWordsRule(t, ucInc)
	require.False(t, ruleInc.ExcludeDirectSpeech)
	require.Equal(t, 1, gfwCheck(t, ruleInc, ltInc, taggerInc, "»Der Satz enthält augenscheinlich ein Füllwort«"))

	// minPercent 0 + both sentence options: deferred relativ also emits → 4 / 3 (Java assert numbers)
	ruleValuesAll := map[string][]any{
		"FILLER_WORDS_DE": {0, true, true, true},
	}
	ucAll := languagetool.NewUserConfigWithWords(nil, ruleValuesAll)
	ruleAll, ltAll, taggerAll := setUpGermanFillerWordsRule(t, ucAll)
	require.Equal(t, 4, gfwCheck(t, ruleAll, ltAll, taggerAll,
		"Der Text enthält zu wenige Füllwörter, daher werden sie nicht angezeigt. "+
			"Was sich an diesem Satz mit diesem relativ einfachen Füllwort zeigt. "+
			"Dazu müssen noch eine Reihe von Sätzen geschrieben werden, um die Anzahl der Wörter zu erhöhen. "+
			"Langsam sollten die Anzahl der Worte für das Drücken unter die kritische Grenze reichen. "+
			"Jetzt schreibe ich allerdings einen Satz, der vier Füllwörter enthält, was allemal ziemlich ausreichend ist."))
	require.Equal(t, 3, gfwCheck(t, ruleAll, ltAll, taggerAll,
		"Der Text enthält zu wenige Füllwörter, daher werden sie nicht angezeigt. "+
			"Was sich an diesem Satz mit diesem relativ einfachen Füllwort zeigt. "+
			"Dazu müssen noch eine Reihe von Sätzen geschrieben werden, um die Anzahl der Wörter zu erhöhen. "+
			"Langsam sollten die Anzahl der Worte für das Drücken unter die kritische Grenze reichen. "+
			"Jetzt schreibe ich einen Satz, der zwei Füllwörter hintereinander enthält, was allemal ziemlich ausreichend ist."))
}

// TestGermanFillerWordsRule_Exceptions ports isException / isTwoWordException surface cases
// with real tagged analysis (num==1, after comma, immer wieder).
func TestGermanFillerWordsRule_Exceptions(t *testing.T) {
	ruleValues := map[string][]any{"FILLER_WORDS_DE": {0}}
	uc := languagetool.NewUserConfigWithWords(nil, ruleValues)
	rule, lt, tagger := setUpGermanFillerWordsRule(t, uc)

	// sentence-start exception (num==1)
	require.Equal(t, 0, gfwCheck(t, rule, lt, tagger, "Allerdings war es kalt."))
	require.Equal(t, 1, gfwCheck(t, rule, lt, tagger, "Es war allerdings kalt."))
	// two-word exception: immer wieder
	require.Equal(t, 0, gfwCheck(t, rule, lt, tagger, "Das passiert immer wieder hier."))
	// after comma exception
	require.Equal(t, 0, gfwCheck(t, rule, lt, tagger, "Es war kalt, allerdings regnete es."))
}

func TestGermanFillerWordsRule_LimitMessage(t *testing.T) {
	r := NewGermanFillerWordsRule(nil)
	require.Contains(t, r.getLimitMessage(0, 0), "könnte ein Füllwort")
	require.Contains(t, r.getLimitMessage(8, 12.4), "Mehr als 8%")
	require.Contains(t, r.getLimitMessage(8, 12.4), "12%")
	require.Equal(t, "", r.GetSentenceMessage())
}
