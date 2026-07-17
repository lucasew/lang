package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreGermanRules installs DE word-repeat + beginning + long-sentence + shared layout.
func RegisterCoreGermanRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "de")
	wr := NewGermanWordRepeatRule(map[string]string{"repetition": "Wortwiederholung"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	wrb := NewGermanWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_adv":       "Drei aufeinanderfolgende Sätze beginnen mit demselben Adverb.",
		"desc_repetition_beginning_word":      "Drei aufeinanderfolgende Sätze beginnen mit demselben Wort.",
		"desc_repetition_beginning_thesaurus": "Erwägen Sie ein Synonym.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	ls := NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "Dieser Satz ist zu lang (%d Wörter)",
	}, 40)
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))

	// Soft surface patterns until full grammar.xml is loaded.
	patterns.RegisterTokenSequences(lt, "de", []patterns.TokenSequenceSpec{
		{ID: "DE_WEGEN_DEM", Tokens: []string{"wegen", "dem"}, Message: "Meinten Sie 'wegen des'?", Suggestion: "wegen des"},
		{ID: "DE_TROTZ_DEM", Tokens: []string{"trotz", "dem"}, Message: "Meinten Sie 'trotz des'?", Suggestion: "trotz des"},
	})

	// Official replace.txt / replace_custom.txt + coherency.txt (vendored/embedded).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))
}
