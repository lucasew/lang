package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreFrenchRules installs shared layout + FR word-repeat + beginning.
func RegisterCoreFrenchRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "fr")
	wr := NewWordRepeatRule(map[string]string{"repetition": "Répétition de mot"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Trois phrases successives commencent par le même mot.",
		"desc_repetition_beginning_adv":  "Trois phrases successives commencent par le même adverbe.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	// Soft surface patterns until full grammar.xml is loaded.
	patterns.RegisterTokenSequences(lt, "fr", []patterns.TokenSequenceSpec{
		{ID: "FR_MALGRE_QUE", Tokens: []string{"malgré", "que"}, Message: "Préférez 'bien que' ou 'quoique'.", Suggestion: "bien que"},
		{ID: "FR_AU_JOURD_HUI", Tokens: []string{"au", "jour", "d", "hui"}, Message: "Écrivez 'aujourd'hui' en un mot.", Suggestion: "aujourd'hui"},
	})

	// Official replace.txt / replace_custom.txt (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	// Official compounds + synonym repeated-words (embedded).
	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	rw := NewFrenchRepeatedWordsRule(nil)
	lt.AddTextLevelRuleChecker(rw.GetID(), rules.AsTextLevelChecker(rw.MatchList))
}
