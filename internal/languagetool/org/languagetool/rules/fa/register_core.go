package fa

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCorePersianRules ports Persian.getRelevantRules.
// Java: PersianWordRepeatRule (PERSIAN_WORD_REPEAT_RULE) +
// PersianWordRepeatBeginningRule (PERSIAN_WORD_REPEAT_BEGINNING_RULE) — not invent FA_ ids.
func RegisterCorePersianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "fa")
	wr := NewPersianWordRepeatRule(map[string]string{"repetition": "تکرار"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewPersianWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "سه جمله پیاپی با یک کلمه آغاز می‌شوند.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	// Official replace + coherency tables (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	// Official space-before rule.
	sb := NewPersianSpaceBeforeRule(nil)
	lt.AddRuleChecker(sb.GetID(), rules.AsSentenceCheckerSimple(sb.Match))
}
