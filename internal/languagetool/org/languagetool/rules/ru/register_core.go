package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCoreRussianRules ports Russian.getRelevantRules (+ chunker / ignored chars).
// Java: RussianSimpleWordRepeatRule (extends WordRepeatRule → WORD_REPEAT_RULE id) +
// RussianWordRepeatRule (RU_WORD_REPEAT) — no WordRepeatBeginning invent.
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
	// Java RussianSimpleWordRepeatRule: default WordRepeatRule id WORD_REPEAT_RULE.
	wr := NewRussianSimpleWordRepeatRule(map[string]string{"repetition": "Повтор слова"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	// Java RussianWordRepeatRule (RU_WORD_REPEAT).
	rwr := NewRussianWordRepeatRule(map[string]string{"repetition": "Повтор слов в предложении"})
	lt.AddRuleChecker(rwr.GetID(), rules.AsSentenceCheckerSimple(rwr.Match))

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
