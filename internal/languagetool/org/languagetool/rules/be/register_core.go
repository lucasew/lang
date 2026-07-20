package be

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreBelarusianRules ports Belarusian.getRelevantRules (+ priority / ignored chars).
// Java has no WordRepeatRule / WordRepeatBeginning — only ParagraphRepeatBeginning via layout.
func RegisterCoreBelarusianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.PriorityForId = language.BelarusianPriorityForId
	// Java Belarusian.getIgnoredCharactersRegex: soft hyphen + combining acute/grave.
	lt.IgnoredCharacters = languagetool.BelarusianIgnoredCharactersRegex
	rules.RegisterSharedLayoutRules(lt, "be")

	// Official replace.txt / specific_case.txt (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	sc := NewBelarusianSpecificCaseRule(nil)
	lt.AddRuleChecker(sc.GetID(), rules.AsSentenceCheckerSimple(sc.Match))

	// Java createDefaultSpellingRule / Morfologik getId; CFSA2 when dict present.
	// Always full Match (fail-closed map Words when binary dict missing; no invent).
	sp := NewMorfologikBelarusianSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
