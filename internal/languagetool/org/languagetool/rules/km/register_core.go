package km

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreKhmerRules installs shared layout + Khmer word-repeat.
func RegisterCoreKhmerRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "km")
	wr := NewKhmerWordRepeatRule(map[string]string{"repetition": "ពាក្យស្ទួន"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
}
