package km

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreKhmerRules ports Khmer.getRelevantRules / createDefaultSpellingRule.
// Java: KhmerWordRepeatRule only — no WordRepeatBeginning.
func RegisterCoreKhmerRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "km")
	wr := NewKhmerWordRepeatRule(map[string]string{"repetition": "ពាក្យស្ទួន"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	// Official replace table (embedded from upstream).
	sr := NewKhmerSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))

	// Official space-before rule.
	sb := NewKhmerSpaceBeforeRule(nil)
	lt.AddRuleChecker(sb.GetID(), rules.AsSentenceCheckerSimple(sb.Match))

	// Java Khmer.getRelevantRules / createDefaultSpellingRule → KhmerHunspellRule (HUNSPELL_RULE).
	// Dict loading deferred; nil dict fails closed (no invent misspell flags).
	sp := NewKhmerHunspellRuleDefault()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
