package it

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreItalianRules installs shared layout + Italian word-repeat.
func RegisterCoreItalianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "it")
	wr := NewItalianWordRepeatRule(map[string]string{"repetition": "Ripetizione di parola"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
}
