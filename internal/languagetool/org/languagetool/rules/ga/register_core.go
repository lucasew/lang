package ga

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreIrishRules installs shared layout + word-repeat (base rule with language id).
func RegisterCoreIrishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "ga")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Athrá"})
	wr.IDOverride = "GA_WORD_REPEAT_RULE"
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
}
