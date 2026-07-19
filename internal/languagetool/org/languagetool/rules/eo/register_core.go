package eo

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreEsperantoRules installs shared layout + Hunspell speller (Java HunspellRule).
func RegisterCoreEsperantoRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "eo")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Ripeto de vorto"})
	// Java WordRepeatRule default id WORD_REPEAT_RULE
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	// Java Esperanto.getRelevantRules / createDefaultSpellingRule → HunspellRule.
	sp := NewEsperantoHunspellRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
