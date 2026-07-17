package fa

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCorePersianRules installs shared layout + word-repeat (base rule with language id).
func RegisterCorePersianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "fa")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "تکرار"})
	wr.IDOverride = "FA_WORD_REPEAT_RULE"
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
}
