package br

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
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
	// Soft invent token sequences removed (faithful-port): incomplete without grammar.xml, not invented.

	// Official topo.txt place-name replace (embedded from upstream).
	tr := NewTopoReplaceRule(nil)
	lt.AddRuleChecker(tr.GetID(), rules.AsSentenceCheckerSimple(tr.Match))

	// Official compounds.txt.
	cr := NewBretonCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))

	// Java createDefaultSpellingRule → MorfologikBretonSpellerRule.
	// Always full Match (IgnoreTaggedWords + hyphen tokenizingPattern).
	sp := NewMorfologikBretonSpellerRule()
	if p := morfologik.DiscoverLanguageDict(MorfologikBretonSpellerRuleDict); p != "" {
		// Binary CFSA2 optional — fail-closed map Words when missing.
		_ = p
	}
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
