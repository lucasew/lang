package it

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCoreItalianRules ports Italian.getRelevantRules / createDefaultSpellingRule.
// Java: ItalianWordRepeatRule only — no WordRepeatBeginning.
func RegisterCoreItalianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "it")
	wr := NewItalianWordRepeatRule(map[string]string{"repetition": "Ripetizione di parola"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	// Java createDefaultSpellingRule → MorfologikItalianSpellerRule.
	// Always register full Match (orderSuggestions capitalization drop); wire filter
	// dict when binary resource is on disk (fail-closed IsMisspelled without it).
	sp := NewMorfologikItalianSpellerRule()
	if p := morfologik.DiscoverLanguageDict(ItalianSpellerDict); p != "" {
		if WireItalianFilterSpeller(p) {
			sp.IsMisspelled = FilterDictIsMisspelled
		}
	}
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
