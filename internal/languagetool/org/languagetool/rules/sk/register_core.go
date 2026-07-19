package sk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreSlovakRules installs shared layout + language word-repeat + beginning.
func RegisterCoreSlovakRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "sk")
	wr := NewWordRepeatRule(map[string]string{"repetition": "Opakovanie"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tri vety po sebe začínajú rovnakým slovom.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
	// Soft invent token sequences removed (faithful-port): incomplete without grammar.xml, not invented.

	// Official compounds.txt (embedded from upstream).
	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))

	// Java createDefaultSpellingRule / Morfologik getId; CFSA2 when dict present.
	// Empty map shell fails closed when binary resource is missing (no invent).
	// Always full Match (fail-closed map Words when binary dict missing; no invent).
	sp := NewMorfologikSlovakSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
