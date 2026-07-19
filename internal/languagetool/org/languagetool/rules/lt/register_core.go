package lt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreLithuanianRules installs shared layout + Morfologik speller.
// Java Lithuanian.getRelevantRules: layout + MorfologikLithuanianSpellerRule (no WordRepeatRule).
func RegisterCoreLithuanianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "lt")

	// Java createDefaultSpellingRule / Morfologik getId; CFSA2 when dict present.
	// Always full Match (fail-closed map Words when binary dict missing; no invent).
	sp := NewMorfologikLithuanianSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
