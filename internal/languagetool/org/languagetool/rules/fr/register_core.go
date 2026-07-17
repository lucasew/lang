package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
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
}
