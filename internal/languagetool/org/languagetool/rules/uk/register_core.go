package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCoreUkrainianRules installs shared layout + Ukrainian word-repeat + beginning.
func RegisterCoreUkrainianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// Java Ukrainian.IGNORED_CHARS: soft hyphen + combining acute.
	lt.IgnoredCharacters = languagetool.UkrainianIgnoredCharactersRegex
	rules.RegisterSharedLayoutRules(lt, "uk")
	wr := NewUkrainianWordRepeatRule(map[string]string{"repetition": "Повтор слова"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Три речення поспіль починаються одним словом.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	// Soft invent token sequences removed (faithful-port): incomplete without grammar.xml, not invented.

	// Official replace tables (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	ss := NewSimpleReplaceSoftRule(nil)
	lt.AddRuleChecker(ss.GetID(), rules.AsSentenceCheckerSimple(ss.Match))
	rn := NewSimpleReplaceRenamedRule(nil)
	lt.AddRuleChecker(rn.GetID(), rules.AsSentenceCheckerSimple(rn.Match))

	// Java createDefaultSpellingRule → MorfologikUkrainianSpellerRule.
	// Always full Match (ignoreToken, filterSuggestions, dash tops, !hasGoodTag).
	sp := NewMorfologikUkrainianSpellerRule()
	if p := morfologik.DiscoverLanguageDict(UkrainianSpellerDict); p != "" {
		if WireUkrainianFilterSpeller(p) {
			// Preserve trailing-hyphen arm: wrap FilterDictIsMisspelledUK.
			inner := FilterDictIsMisspelledUK
			sp.IsMisspelled = func(word string) bool {
				return sp.ukIsMisspelled(word, inner)
			}
		}
	}
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
