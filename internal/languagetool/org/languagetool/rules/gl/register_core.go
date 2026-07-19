package gl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreGalicianRules installs shared layout + language word-repeat + beginning.
func RegisterCoreGalicianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.PriorityForId = language.GalicianPriorityForId
	rules.RegisterSharedLayoutRules(lt, "gl")
	wr := NewWordRepeatRule(map[string]string{"repetition": "Repetición"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tres frases sucesivas comezan coa mesma palabra.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
	// Soft invent token sequences removed (faithful-port): incomplete without grammar.xml, not invented.

	// Official replace.txt (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))

	// Official redundancy + wordiness style tables.
	rd := NewGalicianRedundancyRule(nil)
	lt.AddRuleChecker(rd.GetID(), rules.AsSentenceCheckerSimple(rd.Match))
	pe := NewGalicianWordinessRule(nil)
	lt.AddRuleChecker(pe.GetID(), rules.AsSentenceCheckerSimple(pe.Match))

	// Java Galician.getRelevantRules / createDefaultSpellingRule → HunspellRule (HUNSPELL_RULE).
	sp := NewGalicianHunspellRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
