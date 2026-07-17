package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreCatalanRules installs shared layout + Catalan word-repeat + beginning.
func RegisterCoreCatalanRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "ca")
	wr := NewCatalanWordRepeatRule(map[string]string{"repetition": "Repetició de paraula"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewCatalanWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tres frases successives comencen amb la mateixa paraula.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	patterns.RegisterTokenSequences(lt, "ca", []patterns.TokenSequenceSpec{
		{ID: "CA_A_EL", Tokens: []string{"a", "el"}, Message: "Volíeu dir 'al'?", Suggestion: "al"},
		{ID: "CA_DE_EL", Tokens: []string{"de", "el"}, Message: "Volíeu dir 'del'?", Suggestion: "del"},
	})

	// Official replace + coherency tables (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))
	// Additional official replace tables (anglicisms, multiwords).
	ang := NewSimpleReplaceAnglicism(nil)
	lt.AddRuleChecker(ang.GetID(), rules.AsSentenceCheckerSimple(ang.Match))
	mw := NewSimpleReplaceMultiwordsRule(nil)
	lt.AddRuleChecker(mw.GetID(), rules.AsSentenceCheckerSimple(mw.Match))

	// Official wrong-word, compounds, IEC diacritics, repeated-words synonyms.
	ww := NewCatalanWrongWordInContextRule(nil)
	lt.AddRuleChecker(ww.GetID(), rules.AsSentenceCheckerSimple(ww.Match))
	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	di := NewSimpleReplaceDiacriticsIEC(nil)
	lt.AddRuleChecker(di.GetID(), rules.AsSentenceCheckerSimple(di.Match))
	rw := NewCatalanRepeatedWordsRule(nil)
	lt.AddTextLevelRuleChecker(rw.GetID(), rules.AsTextLevelChecker(rw.MatchList))

	// Additional official CA replace packs (DNV, Balearic, verbs, adverbs, ops).
	dnv := NewSimpleReplaceDNVRule(nil)
	lt.AddRuleChecker(dnv.GetID(), rules.AsSentenceCheckerSimple(dnv.Match))
	dnvC := NewSimpleReplaceDNVColloquialRule(nil)
	lt.AddRuleChecker(dnvC.GetID(), rules.AsSentenceCheckerSimple(dnvC.Match))
	dnvS := NewSimpleReplaceDNVSecondaryRule(nil)
	lt.AddRuleChecker(dnvS.GetID(), rules.AsSentenceCheckerSimple(dnvS.Match))
	bal := NewSimpleReplaceBalearicRule(nil)
	lt.AddRuleChecker(bal.GetID(), rules.AsSentenceCheckerSimple(bal.Match))
	vr := NewSimpleReplaceVerbsRule(nil)
	lt.AddRuleChecker(vr.GetID(), rules.AsSentenceCheckerSimple(vr.Match))
	adv := NewSimpleReplaceAdverbsMent(nil)
	lt.AddRuleChecker(adv.GetID(), rules.AsSentenceCheckerSimple(adv.Match))
	op := NewReplaceOperationNamesRule(nil)
	lt.AddRuleChecker(op.GetID(), rules.AsSentenceCheckerSimple(op.Match))
	cc := NewCheckCaseRule(nil)
	lt.AddRuleChecker(cc.GetID(), rules.AsSentenceCheckerSimple(cc.Match))
	// Valencian coherency variants (official coherency-valencia.txt).
	wcv := NewWordCoherencyValencianRule(nil)
	lt.AddTextLevelRuleChecker(wcv.GetID(), rules.AsTextLevelChecker(wcv.MatchList))
}
