package be

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreBelarusianRules installs shared layout + official replace and specific-case tables.
func RegisterCoreBelarusianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.PriorityForId = language.BelarusianPriorityForId
	// Java Belarusian.getIgnoredCharactersRegex: soft hyphen + combining acute/grave.
	lt.IgnoredCharacters = languagetool.BelarusianIgnoredCharactersRegex
	rules.RegisterSharedLayoutRules(lt, "be")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Паўтор слова"})
	wr.IDOverride = "BE_WORD_REPEAT_RULE"
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := rules.NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Тры сказы запар пачынаюцца адным словам.",
	})
	wrb.IDOverride = "BE_WORD_REPEAT_BEGINNING_RULE"
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	// Official replace.txt / specific_case.txt (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	sc := NewBelarusianSpecificCaseRule(nil)
	lt.AddRuleChecker(sc.GetID(), rules.AsSentenceCheckerSimple(sc.Match))

	// Java createDefaultSpellingRule / Morfologik getId; CFSA2 when dict present.
	// Always full Match (fail-closed map Words when binary dict missing; no invent).
	sp := NewMorfologikBelarusianSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
