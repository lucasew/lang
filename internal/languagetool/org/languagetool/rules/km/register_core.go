package km

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreKhmerRules ports Khmer.getRelevantRules / createDefaultSpellingRule.
// Java list only: KhmerHunspell, SimpleReplace, WordRepeat, UnpairedBrackets,
// SpaceBefore — no invent SharedLayout extras (no comma/whitespace/uppercase/etc.).
func RegisterCoreKhmerRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}

	// Java Khmer.getRelevantRules / createDefaultSpellingRule → KhmerHunspellRule (HUNSPELL_RULE).
	sp := NewKhmerHunspellRuleDefault()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

	// Official replace table (embedded from upstream).
	sr := NewKhmerSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))

	// Java: KhmerWordRepeatRule only — no WordRepeatBeginning.
	wr := NewKhmerWordRepeatRule(map[string]string{"repetition": "ពាក្យស្ទួន"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	// Java KhmerUnpairedBracketsRule (text-level).
	ub := NewKhmerUnpairedBracketsRule(nil)
	lt.AddTextLevelRuleChecker(ub.GetID(), rules.AsTextLevelChecker(ub.MatchList))

	// Official space-before rule.
	sb := NewKhmerSpaceBeforeRule(nil)
	lt.AddRuleChecker(sb.GetID(), rules.AsSentenceCheckerSimple(sb.Match))
}
