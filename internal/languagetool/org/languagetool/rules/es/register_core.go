package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreSpanishRules installs shared layout + Spanish word-repeat + beginning.
func RegisterCoreSpanishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "es")
	wr := NewSpanishWordRepeatRule(map[string]string{"repetition": "Repetición de palabra"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewSpanishWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tres oraciones sucesivas comienzan con la misma palabra.",
		"desc_repetition_beginning_adv":  "Tres oraciones sucesivas comienzan con el mismo adverbio.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
}
