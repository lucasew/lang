package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreSpanishRules installs shared layout + Spanish word-repeat + beginning.
func RegisterCoreSpanishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "es")
	wr := NewSpanishWordRepeatRule(map[string]string{"repetition": "Repetición de palabra"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewSpanishWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tres oraciones sucesivas comienzan con la misma palabra.",
		"desc_repetition_beginning_adv":  "Tres oraciones sucesivas comienzan con el mismo adverbio.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	patterns.RegisterTokenSequences(lt, "es", []patterns.TokenSequenceSpec{
		{ID: "ES_A_EL", Tokens: []string{"a", "el"}, Message: "¿Quiso decir 'al'?", Suggestion: "al"},
		{ID: "ES_DE_EL", Tokens: []string{"de", "el"}, Message: "¿Quiso decir 'del'?", Suggestion: "del"},
	})

	// Official replace.txt / replace_custom.txt (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	// Official compounds + synonym repeated-words (embedded).
	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	rw := NewSpanishRepeatedWordsRule(nil)
	lt.AddTextLevelRuleChecker(rw.GetID(), rules.AsTextLevelChecker(rw.MatchList))

	// Official wrong-word-in-context + verb replace tables.
	ww := NewSpanishWrongWordInContextRule(nil)
	lt.AddRuleChecker(ww.GetID(), rules.AsSentenceCheckerSimple(ww.Match))
	vr := NewSimpleReplaceVerbsRule(nil)
	lt.AddRuleChecker(vr.GetID(), rules.AsSentenceCheckerSimple(vr.Match))
}
