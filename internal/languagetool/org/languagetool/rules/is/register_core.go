package is

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreIcelandicRules installs shared layout + HunspellNoSuggestion speller.
// Java Icelandic.getRelevantRules registers HunspellNoSuggestionRule.
func RegisterCoreIcelandicRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "is")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Endurtekning orðs"})
	wr.IDOverride = "IS_WORD_REPEAT_RULE"
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := rules.NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Þrjár setningar í röð byrja á sama orði.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	sp := NewIcelandicHunspellNoSuggestionRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
