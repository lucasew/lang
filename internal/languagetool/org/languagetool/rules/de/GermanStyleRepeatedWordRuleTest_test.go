package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanStyleRepeatedWordRuleTest.java
//
// Java:
//   Language lang = Languages.getLanguageForShortCode("de-DE");
//   JLanguageTool lt = new JLanguageTool(lang);
//   setUpRule(lt, null | getRuleValues(1, false, true));
//   assertEquals(n, lt.check(text).size());
//
// Go: de-DE analysis via GermanWordTokenizer + GermanTagger.Tag
// (same path as UnnecessaryPhraseRule / NonSignificantVerbsRule twins).
// Compound path wires official /de/hunspell/de_DE.dict (Java MorfologikSpeller).
// No invent morph tables; match counts must equal Java.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
	"github.com/stretchr/testify/require"
)

// setUpGermanStyleRepeatedWordRule ports GermanStyleRepeatedWordRuleTest.setUpRule.
func setUpGermanStyleRepeatedWordRule(t *testing.T, userConfig *languagetool.UserConfig) (*GermanStyleRepeatedWordRule, *languagetool.JLanguageTool, *detag.GermanTagger) {
	t.Helper()
	// Java: Languages.getLanguageForShortCode("de-DE")
	lt := languagetool.NewJLanguageTool("de-DE")
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for GermanStyleRepeatedWordRule twin")
	tagger := DiscoveredGermanTagger()
	require.NotNil(t, tagger, "GermanTagger.INSTANCE resources required")

	// Java: for (Rule rule : lt.getAllRules()) { lt.disableRule(rule.getId()); }
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		lt.DisableRule(id)
	}
	// Java: new GermanStyleRepeatedWordRule(TestTools.getMessages(...), lang, userConfig)
	rule := NewGermanStyleRepeatedWordRuleWithUserConfig(nil, userConfig)
	lt.AddTextLevelRuleChecker(rule.GetID(), rules.AsTextLevelChecker(rule.MatchList))
	lt.EnableRule(rule.GetID())
	return rule, lt, tagger
}

// gsrwCheck ports lt.check(text).size() with only GermanStyleRepeatedWordRule enabled.
func gsrwCheck(t *testing.T, rule *GermanStyleRepeatedWordRule, lt *languagetool.JLanguageTool, tagger *detag.GermanTagger, text string) int {
	t.Helper()
	sents := analyzeTextWithGermanTagger(lt, tagger, text)
	return len(rule.MatchList(sents))
}

// getGermanStyleRuleValues ports GermanStyleRepeatedWordRuleTest.getRuleValues.
func getGermanStyleRuleValues(numSentences int, excludeDirectSpeech, testComposedWords bool) map[string][]any {
	return map[string][]any{
		germanStyleRuleID: {numSentences, excludeDirectSpeech, testComposedWords},
	}
}

func TestGermanStyleRepeatedWordRule_Meta(t *testing.T) {
	rule := NewGermanStyleRepeatedWordRule(nil)
	require.Equal(t, germanStyleRuleID, rule.GetID())
	require.Equal(t, "STYLE_REPEATED_WORD_RULE_DE", rule.GetID())
	require.Equal(t, germanStyleDesc, rule.GetDescription())
	require.True(t, rule.DefaultOff)
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, 1, rule.MaxDistanceOfSentences)
	require.True(t, rule.ExcludeDirectSpeech)
	require.False(t, rule.TestCompoundWords) // Java TEST_COMPOUND_WORDS default false
	require.Equal(t, -1, rule.MinToCheckParagraph())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	require.NotNil(t, rule.GetCategory())
	require.NotEmpty(t, rule.GetIncorrectExamples())
	opts := rule.GetRuleOptions()
	require.Len(t, opts, 3)
	require.Equal(t, 1, opts[0].DefaultValue)
	require.Equal(t, true, opts[1].DefaultValue)
	require.Equal(t, false, opts[2].DefaultValue)
	require.Equal(t, germanStyleOptExcludeSpeech, opts[1].ConfigureText)
	require.Equal(t, germanStyleOptTestCompound, opts[2].ConfigureText)
}

// TestGermanStyleRepeatedWordRule_Rule ports GermanStyleRepeatedWordRuleTest.testRule
// with real de-DE tagging (Java JLanguageTool + GermanTagger).
func TestGermanStyleRepeatedWordRule_Rule(t *testing.T) {
	// Java: setUpRule(lt, null) — defaults: maxDistance=1, excludeDirectSpeech=true, testCompoundWords=false
	rule, lt, tagger := setUpGermanStyleRepeatedWordRule(t, nil)

	// Java assertEquals(2, … großen … großen …)
	require.Equal(t, 2, gsrwCheck(t, rule, lt, tagger,
		"Der alte Mann wohnte in einem großen Haus. Es stand in einem großen Garten."))
	// Java assertEquals(0, …)
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Der alte Mann wohnte in einem großen Haus. Es stand in einem weitläufigen Garten."))
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Endlos lang zog sich der Ton dahin, aber schließlich verklang er doch."))
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Die Au leuchtete im Blitz, die Glühwürmchen flimmerten vor Aufregung."))
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Sie sind alle in der gleichen Art hergestellt. Nur bei einem Artefakt wurde eine andere Vorgehensweise angewandt."))
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Ich erkannte keinen Lidschlag zu spät, dass es nicht in ihrer Absicht lag."))
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Dieser Schritt brachte uns näher an die Lösung. Der nächste würde das Problem für immer aus der Welt schaffen."))
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Er lehnte seinen Wanderstab gegen die Wand."))
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Zweifel lagen in der Luft, ob es richtig war, diesen Weg einzuschlagen."))
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Der Donnerhall verklang nur langsam in meinen Ohren. Das war mir in all den Jahren meines Lebens noch nicht passiert."))
	// default: compound parts off → Schiffsmotor/Motor not flagged
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Der Schiffsmotor, der im Heck des Schiffs eingebaut war, röhrte. Auf Hochtouren lief der Motor."))
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Der Buntspecht stolzierte den Baum hoch. Schon klopfte der Specht."))
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Rotbraun war die Farbe der Haselnuss. Der Horizont schimmerte rot."))
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Rotbraun war die Farbe der Haselnuss. Was für eine 'Farbe'?"))
	// excludeDirectSpeech=true → quoted Ausblick not checked
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Der Ausblick war überwätigend. \"Was für ein Ausblick!\", sagte sie auch prompt."))
	// Frau + EIG exception
	require.Equal(t, 0, gsrwCheck(t, rule, lt, tagger,
		"Frau Heinrich stand unschlüssig in in ihrem Garten. Frau Müller schaute zu ihr herüber."))

	// Java: setUpRule(lt, getRuleValues(1, false, true))
	// → maxDistance=1, excludeDirectSpeech=false, testCompoundWords=true
	// Compound-part matching needs Java Morfologik /de/hunspell/de_DE.dict.
	dict := deHunspellPath(t, "de_DE.dict")
	require.True(t, WireGermanFilterSpeller(dict), "de_DE.dict required for compound-word twin (Java MorfologikSpeller)")
	t.Cleanup(ClearGermanFilterSpeller)

	uc := languagetool.NewUserConfigWithWords(nil, getGermanStyleRuleValues(1, false, true))
	rule2, lt2, tagger2 := setUpGermanStyleRepeatedWordRule(t, uc)
	require.Equal(t, 1, rule2.MaxDistanceOfSentences)
	require.False(t, rule2.ExcludeDirectSpeech)
	require.True(t, rule2.TestCompoundWords)
	require.Equal(t, 1, rule2.MinToCheckParagraph()) // excludeDirectSpeech false → maxDistance

	// Java: Schiffsmotor/Schiffs same-sentence + Motor → 3
	require.Equal(t, 3, gsrwCheck(t, rule2, lt2, tagger2,
		"Der Schiffsmotor, der im Heck des Schiffs eingebaut war, röhrte. Auf Hochtouren lief der Motor."))
	// Java: Buntspecht … Specht → 2
	require.Equal(t, 2, gsrwCheck(t, rule2, lt2, tagger2,
		"Der Buntspecht stolzierte den Baum hoch. Schon klopfte der Specht."))
	// Java: Rotbraun … rot → 2
	require.Equal(t, 2, gsrwCheck(t, rule2, lt2, tagger2,
		"Rotbraun war die Farbe der Haselnuss. Der Horizont schimmerte rot."))
	// Java: excludeDirectSpeech=false → quoted Ausblick flagged → 2
	require.Equal(t, 2, gsrwCheck(t, rule2, lt2, tagger2,
		"Der Ausblick war überwätigend. \"Was für ein Ausblick!\", sagte sie auch prompt."))
}

func TestGermanStyleRepeatedWordRule_URLOnMatch(t *testing.T) {
	// Java setURL: single lemma → OpenThesaurus + lemma; match gets URL.
	rule, lt, tagger := setUpGermanStyleRepeatedWordRule(t, nil)
	text := "Der alte Mann wohnte in einem großen Haus. Es stand in einem großen Garten."
	sents := analyzeTextWithGermanTagger(lt, tagger, text)
	ms := rule.MatchList(sents)
	require.Len(t, ms, 2)
	for _, m := range ms {
		require.NotEmpty(t, m.GetURL(), "Java setURL attaches OpenThesaurus URL")
		require.Contains(t, m.GetURL(), "openthesaurus.de/synonyme/")
	}
}

func TestGermanStyleRepeatedWordRule_IsUnknownWord(t *testing.T) {
	// Java isUnknownWord: isPosTagUnknown && len>2 && letters only.
	unk := languagetool.AnalyzePlain("Blahxyz").GetTokensWithoutWhitespace()
	var blah *languagetool.AnalyzedTokenReadings
	for _, tok := range unk {
		if tok != nil && tok.GetToken() == "Blahxyz" {
			blah = tok
			break
		}
	}
	require.NotNil(t, blah)
	require.True(t, blah.IsPosTagUnknown())
	require.True(t, isUnknownWordStyle(blah))
	// tagged word is not "unknown"
	_, lt, tagger := setUpGermanStyleRepeatedWordRule(t, nil)
	sents := analyzeTextWithGermanTagger(lt, tagger, "Haus.")
	var haus *languagetool.AnalyzedTokenReadings
	for _, tok := range sents[0].GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "Haus" {
			haus = tok
			break
		}
	}
	require.NotNil(t, haus)
	require.False(t, isUnknownWordStyle(haus))
}
