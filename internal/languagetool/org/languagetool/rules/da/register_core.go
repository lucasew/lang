package da

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreDanishRules installs shared layout + language word-repeat + beginning.
func RegisterCoreDanishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "da")
	wr := NewWordRepeatRule(map[string]string{"repetition": "Gentagelse"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tre sætninger i træk begynder med samme ord.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	// Soft invent token sequences removed (faithful-port): incomplete without grammar.xml, not invented.

	// Java Danish.getRelevantRules / createDefaultSpellingRule → HunspellRule (HUNSPELL_RULE).
	// Dict loading deferred; nil dict fails closed (no invent misspell flags).
	sp := NewDanishHunspellRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
