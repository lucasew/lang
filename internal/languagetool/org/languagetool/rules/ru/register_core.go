package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreRussianRules installs shared layout + Russian word-repeat + beginning.
func RegisterCoreRussianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "ru")
	// Simple adjacent is more reliable without full POS; advanced RU needs lemmas.
	wr := NewRussianSimpleWordRepeatRule(map[string]string{"repetition": "Повтор слова"})
	if wr.WordRepeatRule != nil && wr.IDOverride == "" {
		wr.IDOverride = "RU_WORD_REPEAT_SIMPLE"
	}
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Три подряд идущих предложения начинаются с одного слова.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	patterns.RegisterTokenSequences(lt, "ru", []patterns.TokenSequenceSpec{
		{ID: "RU_В_В", Tokens: []string{"в", "в"}, Message: "Возможный повтор предлога «в».", Suggestion: "в"},
		{ID: "RU_И_И", Tokens: []string{"и", "и"}, Message: "Возможный повтор союза «и».", Suggestion: "и"},
	})

	// Official replace + coherency tables (embedded from upstream).
	sr := NewRussianSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	wc := NewRussianWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	// Official compounds, dash compounds, specific-case.
	cr := NewRussianCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	dr := NewRussianDashRule(nil)
	lt.AddRuleChecker(dr.GetID(), rules.AsSentenceCheckerSimple(dr.Match))
	sc := NewRussianSpecificCaseRule(nil)
	lt.AddRuleChecker(sc.GetID(), rules.AsSentenceCheckerSimple(sc.Match))
}
