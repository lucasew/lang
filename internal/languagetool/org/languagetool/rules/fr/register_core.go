package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// languagetool.TokenTag used by late-bound TagPOS.

// RegisterCoreFrenchRules ports French.getRelevantRules / createDefaultSpellingRule.
// Java has no WordRepeatRule / WordRepeatBeginning — FrenchRepeatedWordsRule is the
// FR-specific repetition style rule (not invent generic WR).
func RegisterCoreFrenchRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// Java French.getPriorityForId + filterRuleMatches. Hooks set by language init
	// — avoid import cycle (language tests import rules/fr).
	if languagetool.FrenchPriorityForIdHook != nil {
		lt.PriorityForId = languagetool.FrenchPriorityForIdHook
	}
	if languagetool.FilterFrenchRuleMatchesHook != nil {
		lt.FilterRuleMatches = languagetool.FilterFrenchRuleMatchesHook
	}
	rules.RegisterSharedLayoutRules(lt, "fr")

	// Official replace.txt / replace_custom.txt (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	// Official compounds + synonym repeated-words (embedded).
	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	rw := NewFrenchRepeatedWordsRule(nil)
	lt.AddTextLevelRuleChecker(rw.GetID(), rules.AsTextLevelChecker(rw.MatchList))

	// Java createDefaultSpellingRule → MorfologikFrenchSpellerRule (FR_SPELLING_RULE).
	// Always register full Match (orderSuggestions + apostrophe/hyphen tops via TagPOS).
	sp := NewMorfologikFrenchSpellerRule()
	if p := morfologik.DiscoverLanguageDict(FrenchSpellerDict); p != "" {
		if WireFrenchFilterSpeller(p) {
			sp.IsMisspelled = FilterDictIsMisspelled
		}
	}
	// Late-bind TagPOS to lt.TagWord: core_rules_checker installs POS dict after RegisterCore.
	// Fail-closed until TagWord is set (no invent POS).
	ltRef := lt
	WireFrenchSpellerTagPOS(sp, func(token string) []languagetool.TokenTag {
		if ltRef.TagWord == nil {
			return nil
		}
		return ltRef.TagWord(token)
	})
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
