package gl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreGalicianRules ports Galician.getRelevantRules / createDefaultSpellingRule.
// Java has no WordRepeatRule / WordRepeatBeginning (ParagraphRepeatBeginning via layout).
func RegisterCoreGalicianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.PriorityForId = language.GalicianPriorityForId
	rules.RegisterSharedLayoutRules(lt, "gl")

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
