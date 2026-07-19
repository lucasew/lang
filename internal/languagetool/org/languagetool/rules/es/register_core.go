package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCoreSpanishRules installs shared layout + Spanish word-repeat + beginning.
func RegisterCoreSpanishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// Java Spanish.getPriorityForId on Check priorities.
	lt.PriorityForId = language.SpanishPriorityForId
	// Java Spanish.adaptSuggestion for AdaptSuggestionsFilter (pattern rules).
	languagetool.LanguageAdaptSuggestionByCode["es"] = language.SpanishAdaptSuggestion
	// Java Spanish.filterRuleMatches (AI_ES_GGEC). Hook from language init (cycle-safe).
	if languagetool.FilterSpanishRuleMatchesHook != nil {
		lt.FilterRuleMatches = languagetool.FilterSpanishRuleMatchesHook
	}
	// Java getTagger() for voseo suggestion drop — wire POS dict when available
	// (fail-closed empty tagger if spanish.dict not on disk).
	_ = language.TryWireSpanishVoseoWordTagger()
	rules.RegisterSharedLayoutRules(lt, "es")
	wr := NewSpanishWordRepeatRule(map[string]string{"repetition": "Repetición de palabra"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewSpanishWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tres oraciones sucesivas comienzan con la misma palabra.",
		"desc_repetition_beginning_adv":  "Tres oraciones sucesivas comienzan con el mismo adverbio.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	// Soft invent token sequences removed (faithful-port): incomplete without grammar.xml, not invented.

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

	// Java createDefaultSpellingRule → MorfologikSpanishSpellerRule.
	// Always register full Match (orderSuggestions + pronoun/digit tops via TagPOS).
	sp := NewMorfologikSpanishSpellerRule()
	if p := morfologik.DiscoverLanguageDict(SpanishSpellerDict); p != "" {
		if WireSpanishFilterSpeller(p) {
			sp.IsMisspelled = FilterDictIsMisspelled
		}
	}
	// Late-bind TagPOS to lt.TagWord (POS may be installed after RegisterCore).
	ltRef := lt
	WireSpanishSpellerTagPOS(sp, func(token string) []languagetool.TokenTag {
		if ltRef.TagWord == nil {
			return nil
		}
		return ltRef.TagWord(token)
	})
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
