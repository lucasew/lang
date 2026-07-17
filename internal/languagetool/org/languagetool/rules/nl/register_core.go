package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreDutchRules installs shared layout + NL word-repeat + beginning.
func RegisterCoreDutchRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "nl")
	wr := NewWordRepeatRule(map[string]string{"repetition": "Woordherhaling"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Drie opeenvolgende zinnen beginnen met hetzelfde woord.",
		"desc_repetition_beginning_adv":  "Drie opeenvolgende zinnen beginnen met hetzelfde bijwoord.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
}
