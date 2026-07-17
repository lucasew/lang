package km

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreKhmerRules installs shared layout + Khmer word-repeat + beginning.
func RegisterCoreKhmerRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "km")
	wr := NewKhmerWordRepeatRule(map[string]string{"repetition": "ពាក្យស្ទួន"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "ប្រយោគបីជាប់គ្នាចាប់ផ្តើមដោយពាក្យដូចគ្នា។",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	// Official replace table (embedded from upstream).
	sr := NewKhmerSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
}
