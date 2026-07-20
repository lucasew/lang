package is

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreIcelandicRules ports Icelandic.getRelevantRules / createDefaultSpellingRule.
// Java: layout + HunspellNoSuggestionRule + WordRepeatRule — no WordRepeatBeginning.
func RegisterCoreIcelandicRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "is")
	// Java WordRepeatRule default id is WORD_REPEAT_RULE (no invent IS_ prefix).
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Endurtekning orðs"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	sp := NewIcelandicHunspellNoSuggestionRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
