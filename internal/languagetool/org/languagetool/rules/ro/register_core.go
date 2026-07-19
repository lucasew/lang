package ro

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreRomanianRules installs shared layout + language word-repeat + beginning.
func RegisterCoreRomanianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "ro")
	wr := NewWordRepeatRule(map[string]string{"repetition": "Repetiție"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewRomanianWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Trei propoziții consecutive încep cu același cuvânt.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
	// Soft invent token sequences removed (faithful-port): incomplete without grammar.xml, not invented.

	// Official replace.txt (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))

	// Official compounds.txt.
	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))

	// Java createDefaultSpellingRule / Morfologik getId; CFSA2 when dict present.
	// Empty map shell fails closed when binary resource is missing (no invent).
	// Always full Match (fail-closed map Words when binary dict missing; no invent).
	sp := NewMorfologikRomanianSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
