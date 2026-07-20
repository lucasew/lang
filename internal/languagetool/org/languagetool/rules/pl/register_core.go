package pl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCorePolishRules ports Polish.getRelevantRules / createDefaultSpellingRule.
// Java: WordRepeatRule (WORD_REPEAT_RULE) + PolishWordRepeatRule (PL_WORD_REPEAT) —
// no WordRepeatBeginning.
func RegisterCorePolishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.PriorityForId = language.PolishPriorityForId
	rules.RegisterSharedLayoutRules(lt, "pl")
	// Java order: WordRepeatRule then later PolishWordRepeatRule (Advanced).
	gwr := rules.NewWordRepeatRule(map[string]string{"repetition": "Powtórzenie"})
	lt.AddRuleChecker(gwr.GetID(), rules.AsSentenceCheckerSimple(gwr.Match))
	pwr := NewPolishWordRepeatRule(map[string]string{"repetition": "Powtórzenie"})
	lt.AddRuleChecker(pwr.GetID(), rules.AsSentenceCheckerSimple(pwr.Match))

	// Official replace + coherency tables (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	// Official compounds + dash compounds.
	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	dr := NewDashRule(nil)
	lt.AddRuleChecker(dr.GetID(), rules.AsSentenceCheckerSimple(dr.Match))

	// Java createDefaultSpellingRule → MorfologikPolishSpellerRule.
	// Always register full Match (isNotCompound, pruneSuggestions, niby-/quasi-);
	// wire filter dict when binary present (fail-closed IsMisspelled without it).
	sp := NewMorfologikPolishSpellerRule()
	if p := morfologik.DiscoverLanguageDict(PolishSpellerDict); p != "" {
		if WirePolishFilterSpeller(p) {
			sp.IsMisspelled = FilterDictIsMisspelled
		}
	}
	ltRef := lt
	WirePolishSpellerTagPOS(sp, func(token string) []languagetool.TokenTag {
		if ltRef.TagWord == nil {
			return nil
		}
		return ltRef.TagWord(token)
	})
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
