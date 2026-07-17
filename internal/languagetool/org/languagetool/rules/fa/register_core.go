package fa

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
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
	patterns.RegisterTokenSequences(lt, "fa", []patterns.TokenSequenceSpec{
		{ID: "FA_در_در", Tokens: []string{"در", "در"}, Message: "تکرار احتمالی «در».", Suggestion: "در"},
	})

	// Official replace + coherency tables (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	// Official space-before rule.
	sb := NewPersianSpaceBeforeRule(nil)
	lt.AddRuleChecker(sb.GetID(), rules.AsSentenceCheckerSimple(sb.Match))
}
