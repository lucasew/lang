package el

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreGreekRules ports Greek.getRelevantRules / createDefaultSpellingRule.
// Java: WordRepeatRule (default WORD_REPEAT_RULE) + GreekWordRepeatBeginningRule.
func RegisterCoreGreekRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "el")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Επανάληψη"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewGreekWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Τρεις διαδοχικές προτάσεις αρχίζουν με την ίδια λέξη.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	// Official homonyms replace table (embedded from upstream).
	hr := NewReplaceHomonymsRule(nil)
	lt.AddRuleChecker(hr.GetID(), rules.AsSentenceCheckerSimple(hr.Match))

	// Official redundancy + specific-case tables.
	rd := NewGreekRedundancyRule(nil)
	lt.AddRuleChecker(rd.GetID(), rules.AsSentenceCheckerSimple(rd.Match))
	sc := NewGreekSpecificCaseRule(nil)
	lt.AddRuleChecker(sc.GetID(), rules.AsSentenceCheckerSimple(sc.Match))

	// Java createDefaultSpellingRule / Morfologik getId; CFSA2 when dict present.
	// Empty map shell fails closed when binary resource is missing (no invent).
	// Always full Match (fail-closed map Words when binary dict missing; no invent).
	sp := NewMorfologikGreekSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
