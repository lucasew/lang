package sl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreSlovenianRules installs shared layout + Slovenian word-repeat + beginning.
func RegisterCoreSlovenianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "sl")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Ponovitev besede"})
	wr.IDOverride = "SL_WORD_REPEAT_RULE"
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tri zaporedne povedi se začnejo z isto besedo.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
	// Soft invent token sequences removed (faithful-port): incomplete without grammar.xml, not invented.

	// Java createDefaultSpellingRule / Morfologik getId; CFSA2 when dict present.
	// Empty map shell fails closed when binary resource is missing (no invent).
	// Always full Match (fail-closed map Words when binary dict missing; no invent).
	sp := NewMorfologikSlovenianSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
