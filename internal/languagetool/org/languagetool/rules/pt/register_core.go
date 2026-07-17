package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCorePortugueseRules installs shared layout + Portuguese word-repeat + beginning.
func RegisterCorePortugueseRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "pt")
	wr := NewPortugueseWordRepeatRule(map[string]string{"repetition": "Repetição de palavra"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewPortugueseWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Três frases sucessivas começam com a mesma palavra.",
		"desc_repetition_beginning_adv":  "Três frases sucessivas começam com o mesmo advérbio.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
}
