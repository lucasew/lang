package da

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreDanishRules ports Danish.getRelevantRules / createDefaultSpellingRule.
// Java: layout + HunspellRule only. Comment in Danish.java: WORD_REPEAT_RULE is in
// grammar.xml — do not invent a class-based WordRepeat / WordRepeatBeginning.
func RegisterCoreDanishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "da")

	// Java Danish.getRelevantRules / createDefaultSpellingRule → HunspellRule (HUNSPELL_RULE).
	// Dict loading deferred; nil dict fails closed (no invent misspell flags).
	sp := NewDanishHunspellRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
