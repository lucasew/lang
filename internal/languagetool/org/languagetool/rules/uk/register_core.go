package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreUkrainianRules installs shared layout + Ukrainian word-repeat + beginning.
func RegisterCoreUkrainianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "uk")
	wr := NewUkrainianWordRepeatRule(map[string]string{"repetition": "Повтор слова"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Три речення поспіль починаються одним словом.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	patterns.RegisterTokenSequences(lt, "uk", []patterns.TokenSequenceSpec{
		{ID: "UK_В_В", Tokens: []string{"в", "в"}, Message: "Можливий повтор прийменника «в».", Suggestion: "в"},
		{ID: "UK_З_З", Tokens: []string{"з", "з"}, Message: "Можливий повтор прийменника «з».", Suggestion: "з"},
	})

	// Official replace tables (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	ss := NewSimpleReplaceSoftRule(nil)
	lt.AddRuleChecker(ss.GetID(), rules.AsSentenceCheckerSimple(ss.Match))
	rn := NewSimpleReplaceRenamedRule(nil)
	lt.AddRuleChecker(rn.GetID(), rules.AsSentenceCheckerSimple(rn.Match))
}
