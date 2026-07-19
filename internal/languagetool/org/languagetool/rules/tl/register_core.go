package tl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreTagalogRules installs shared layout + Morfologik speller.
// Java Tagalog.getRelevantRules: layout + MorfologikTagalogSpellerRule (no WordRepeatRule).
func RegisterCoreTagalogRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "tl")

	// Java createDefaultSpellingRule / Morfologik getId; CFSA2 when dict present.
	// Empty map shell fails closed when binary resource is missing (no invent).
	// Always full Match (fail-closed map Words when binary dict missing; no invent).
	sp := NewMorfologikTagalogSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
