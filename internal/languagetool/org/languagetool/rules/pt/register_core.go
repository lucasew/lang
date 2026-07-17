package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
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

	patterns.RegisterTokenSequences(lt, "pt", []patterns.TokenSequenceSpec{
		{ID: "PT_A_O", Tokens: []string{"a", "o"}, Message: "Talvez 'ao'?", Suggestion: "ao"},
		{ID: "PT_DE_O", Tokens: []string{"de", "o"}, Message: "Talvez 'do'?", Suggestion: "do"},
	})
}
