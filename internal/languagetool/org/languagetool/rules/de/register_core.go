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
	// Compounds + synonym-based repeated words + compound-form coherency (official data).
	cr := NewGermanCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	rw := NewGermanRepeatedWordsRule(nil)
	lt.AddTextLevelRuleChecker(rw.GetID(), rules.AsTextLevelChecker(rw.MatchList))
	cc := NewCompoundCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(cc.GetID(), rules.AsTextLevelChecker(cc.MatchList))

	// Official wrong-word-in-context + dash compounds.
	ww := NewGermanWrongWordInContextRule(nil)
	lt.AddRuleChecker(ww.GetID(), rules.AsSentenceCheckerSimple(ww.Match))
	dr := NewDashRule(nil)
	lt.AddRuleChecker(dr.GetID(), rules.AsSentenceCheckerSimple(dr.Match))

	// Case rule, Swiss compounds, metric unit conversion (official ports).
	cas := NewCaseRule(nil)
	lt.AddRuleChecker(cas.GetID(), rules.AsSentenceCheckerSimple(cas.Match))
	sw := NewSwissCompoundRule(nil)
	lt.AddRuleChecker(sw.GetID(), rules.AsSentenceCheckerSimple(sw.Match))
	uc := NewUnitConversionRule(nil)
	lt.AddRuleChecker(uc.GetID(), rules.AsSentenceCheckerSimple(uc.Match))
}
