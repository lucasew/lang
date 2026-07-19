package ast

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreAsturianRules installs shared layout + Morfologik speller.
// Java Asturian.getRelevantRules: layout + MorfologikAsturianSpellerRule (no WordRepeatRule).
func RegisterCoreAsturianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "ast")

	// Java Asturian.getRelevantRules / createDefaultSpellingRule → MorfologikAsturianSpellerRule.
	// Dict loading deferred; nil dict fails closed (no invent misspell flags).
	sp := NewMorfologikAsturianSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
