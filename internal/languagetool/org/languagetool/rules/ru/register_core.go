package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCoreRussianRules installs shared layout + Russian word-repeat + beginning.
func RegisterCoreRussianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// createDefaultPostDisambiguationChunker → RussianChunker
	WireRussianChunker(lt)
	lt.PriorityForId = language.RussianPriorityForId
	// Java Russian.getIgnoredCharactersRegex: soft hyphen + combining acute/grave.
	lt.IgnoredCharacters = languagetool.RussianIgnoredCharactersRegex
	rules.RegisterSharedLayoutRules(lt, "ru")
	// Simple adjacent is more reliable without full POS; advanced RU needs lemmas.
	wr := NewRussianSimpleWordRepeatRule(map[string]string{"repetition": "Повтор слова"})
	if wr.WordRepeatRule != nil && wr.IDOverride == "" {
		wr.IDOverride = "RU_WORD_REPEAT_SIMPLE"
	}
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Три подряд идущих предложения начинаются с одного слова.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	// Soft invent token sequences removed (faithful-port): incomplete without grammar.xml, not invented.

	// Official replace + coherency tables (embedded from upstream).
	sr := NewRussianSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	wc := NewRussianWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	// Official compounds, dash compounds, specific-case.
	cr := NewRussianCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	dr := NewRussianDashRule(nil)
	lt.AddRuleChecker(dr.GetID(), rules.AsSentenceCheckerSimple(dr.Match))
	sc := NewRussianSpecificCaseRule(nil)
	lt.AddRuleChecker(sc.GetID(), rules.AsSentenceCheckerSimple(sc.Match))

	// Java Russian.createDefaultSpellingRule → MorfologikRussianSpellerRule.
	// Always full Match (ignoreToken letter-gate + filterNoSuggestWords).
	sp := NewMorfologikRussianSpellerRule()
	if p := morfologik.DiscoverLanguageDict(RussianSpellerDict); p != "" {
		if WireRussianFilterSpeller(p) {
			sp.IsMisspelled = FilterDictIsMisspelled
		}
	}
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

	// Java MorfologikRussianYOSpellerRule setDefaultOff (ё-only experimental).
	yo := NewMorfologikRussianYOSpellerRule()
	if p := morfologik.DiscoverLanguageDict(RussianYOSpellerDict); p != "" {
		// YO uses its own dict path; do not overwrite main RU filter wire.
		// IsMisspelled stays map fail-closed unless Words loaded; incomplete without
		// separate YO filter process state (no invent shared filter for yo).
		_ = p
	}
	lt.AddRuleChecker(yo.GetID(), rules.AsSentenceChecker(yo.Match))
	lt.MarkDefaultOff(MorfologikRussianYOSpellerRuleID)
}
