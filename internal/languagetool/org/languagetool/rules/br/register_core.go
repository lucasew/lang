package br

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreBretonRules installs shared layout + word-repeat + beginning.
func RegisterCoreBretonRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "br")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Adlask"})
	wr.IDOverride = "BR_WORD_REPEAT_RULE"
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Teir frazenn war-lerc'h a grog gant ar ger memes.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
	patterns.RegisterTokenSequences(lt, "br", []patterns.TokenSequenceSpec{
		{ID: "BR_HA_HA", Tokens: []string{"ha", "ha"}, Message: "Adlask posubl eus 'ha'.", Suggestion: "ha"},
	})

	// Official topo.txt place-name replace (embedded from upstream).
	tr := NewTopoReplaceRule(nil)
	lt.AddRuleChecker(tr.GetID(), rules.AsSentenceCheckerSimple(tr.Match))
}
