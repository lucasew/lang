package fa

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCorePersianRules installs shared layout + word-repeat + beginning.
func RegisterCorePersianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "fa")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "تکرار"})
	wr.IDOverride = "FA_WORD_REPEAT_RULE"
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewPersianWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "سه جمله پیاپی با یک کلمه آغاز می‌شوند.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
	// Soft invent token sequences removed (faithful-port): incomplete without grammar.xml, not invented.

	// Official replace + coherency tables (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	// Official space-before rule.
	sb := NewPersianSpaceBeforeRule(nil)
	lt.AddRuleChecker(sb.GetID(), rules.AsSentenceCheckerSimple(sb.Match))
}
