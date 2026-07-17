package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreUkrainianRules installs shared layout + Ukrainian word-repeat.
func RegisterCoreUkrainianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "uk")
	wr := NewUkrainianWordRepeatRule(map[string]string{"repetition": "Повтор слова"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
}
