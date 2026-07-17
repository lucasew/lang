package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreRussianRules installs shared layout + Russian word-repeat (simple adjacent).
func RegisterCoreRussianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "ru")
	// Simple adjacent is more reliable without full POS; advanced RU needs lemmas.
	wr := NewRussianSimpleWordRepeatRule(map[string]string{"repetition": "Повтор слова"})
	if wr.WordRepeatRule != nil && wr.IDOverride == "" {
		wr.IDOverride = "RU_WORD_REPEAT_SIMPLE"
	}
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
}
