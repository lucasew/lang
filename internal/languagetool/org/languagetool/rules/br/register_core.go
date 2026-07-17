package br

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreBretonRules installs shared layout + word-repeat (base rule with language id).
func RegisterCoreBretonRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "br")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Adlask"})
	wr.IDOverride = "BR_WORD_REPEAT_RULE"
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
}
