package sk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
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
	patterns.RegisterTokenSequences(lt, "sk", []patterns.TokenSequenceSpec{
		{ID: "SK_V_V", Tokens: []string{"v", "v"}, Message: "Možné opakovanie predložky 'v'.", Suggestion: "v"},
	})
}
