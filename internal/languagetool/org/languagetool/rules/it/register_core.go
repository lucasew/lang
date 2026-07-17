package it

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreItalianRules installs shared layout + Italian word-repeat + beginning.
func RegisterCoreItalianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "it")
	wr := NewItalianWordRepeatRule(map[string]string{"repetition": "Ripetizione di parola"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tre frasi successive iniziano con la stessa parola.",
		"desc_repetition_beginning_adv":  "Tre frasi successive iniziano con lo stesso avverbio.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	patterns.RegisterTokenSequences(lt, "it", []patterns.TokenSequenceSpec{
		{ID: "IT_A_IL", Tokens: []string{"a", "il"}, Message: "Forse intendeva 'al'?", Suggestion: "al"},
		{ID: "IT_DI_IL", Tokens: []string{"di", "il"}, Message: "Forse intendeva 'del'?", Suggestion: "del"},
	})
}
