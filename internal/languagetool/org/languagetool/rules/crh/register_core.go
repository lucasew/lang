package crh

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreCrimeanTatarRules installs shared layout + Morfologik speller.
// Java CrimeanTatar.getRelevantRules: layout + MorfologikCrimeanTatarSpellerRule (no WordRepeatRule).
func RegisterCoreCrimeanTatarRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "crh")

	// Java CrimeanTatar.getRelevantRules / MorfologikCrimeanTatarSpellerRule.
	// Dict loading deferred; nil dict fails closed (no invent misspell flags).
	sp := NewMorfologikCrimeanTatarSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
