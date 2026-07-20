package sv

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreSwedishRules ports Swedish.getRelevantRules / createDefaultSpellingRule.
// Java: WordRepeatRule (default WORD_REPEAT_RULE id) — no WordRepeatBeginning.
func RegisterCoreSwedishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "sv")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Ordrepetition"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	// Official coherency.txt (embedded from upstream).
	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	// Official compounds.txt.
	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))

	// Java Swedish.getRelevantRules / createDefaultSpellingRule → HunspellRule (HUNSPELL_RULE).
	sp := NewSwedishHunspellRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
