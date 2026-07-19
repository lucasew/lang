package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// languagetool.TokenTag used by late-bound TagPOS.

// RegisterCoreFrenchRules installs shared layout + FR word-repeat + beginning.
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
	wr := NewWordRepeatRule(map[string]string{"repetition": "Répétition de mot"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Trois phrases successives commencent par le même mot.",
		"desc_repetition_beginning_adv":  "Trois phrases successives commencent par le même adverbe.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	// Soft invent token sequences removed (faithful-port): use official grammar.xml via LANG_USE_UPSTREAM_GRAMMAR, do not invent surface packs.

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
