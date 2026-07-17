package pl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCorePolishRules installs shared layout + Polish word-repeat + beginning.
func RegisterCorePolishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "pl")
	wr := NewPolishWordRepeatRule(map[string]string{"repetition": "Powtórzenie"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Trzy kolejne zdania zaczynają się od tego samego słowa.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	patterns.RegisterTokenSequences(lt, "pl", []patterns.TokenSequenceSpec{
		{ID: "PL_W_W", Tokens: []string{"w", "w"}, Message: "Możliwe powtórzenie przyimka 'w'.", Suggestion: "w"},
		{ID: "PL_Z_Z", Tokens: []string{"z", "z"}, Message: "Możliwe powtórzenie przyimka 'z'.", Suggestion: "z"},
	})
}
