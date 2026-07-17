package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCorePortugueseRules installs shared layout + Portuguese word-repeat.
func RegisterCorePortugueseRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "pt")
	wr := NewPortugueseWordRepeatRule(map[string]string{"repetition": "Repetição de palavra"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
}
