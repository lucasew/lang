package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// isValencian reports ValencianCatalan language codes (ca-ES-valencia).
func isValencian(langCode string) bool {
	return strings.Contains(strings.ToLower(langCode), "valencia")
}

// RegisterCoreCatalanRules ports Catalan.getRelevantRules / createDefaultSpellingRule.
// Java list only — no invent SharedLayout extras; no CatalanRepeatedWordsRule
// (commented out in Java); Valencian coherency only for Valencian variants.
func RegisterCoreCatalanRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.PriorityForId = language.CatalanPriorityForId
	lt.DefaultRulePriorityForStyle = -50
	languagetool.LanguageAdaptSuggestionByCode["ca"] = language.CatalanAdaptSuggestion
	if languagetool.FilterCatalanRuleMatchesHook != nil {
		lt.FilterRuleMatches = languagetool.FilterCatalanRuleMatchesHook
	}
	if languagetool.FilterCatalanRuleMatchesAfterOverlappingHook != nil {
		lt.FilterRuleMatchesAfterOverlapping = languagetool.FilterCatalanRuleMatchesAfterOverlappingHook
	}

	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	// Java CatalanUnpairedBracketsRule getId override commented → UNPAIRED_BRACKETS.
	ub := NewCatalanUnpairedBracketsRule(nil)
	lt.AddTextLevelRuleChecker(ub.GetID(), rules.AsTextLevelChecker(ub.MatchList))

	up := rules.NewUppercaseSentenceStartRule(nil, "ca")
	lt.AddRuleChecker(up.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	// Java LongSentenceRule(messages, userConfig, 60) → TOO_LONG_SENTENCE.
	ls := rules.NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "This sentence is too long (%d words)",
	}, 60)
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))

	// specific to Catalan (Java order)
	wr := NewCatalanWordRepeatRule(map[string]string{"repetition": "Repetició de paraula"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	sp := NewMorfologikCatalanSpellerRule()
	if p := morfologik.DiscoverLanguageDict(CatalanSpellerDict); p != "" {
		if WireCatalanFilterSpeller(p) {
			sp.IsMisspelled = FilterDictIsMisspelled
		}
	}
	ltRef := lt
	WireCatalanSpellerTagPOS(sp, func(token string) []languagetool.TokenTag {
		if ltRef.TagWord == nil {
			return nil
		}
		return ltRef.TagWord(token)
	})
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

	uq := NewCatalanUnpairedQuestionMarksRule(nil)
	lt.AddTextLevelRuleChecker(uq.GetID(), rules.AsTextLevelChecker(uq.MatchList))
	if uq.IsDefaultOff() {
		lt.MarkDefaultOff(uq.GetID())
	}
	ue := NewCatalanUnpairedExclamationMarksRule(nil)
	lt.AddTextLevelRuleChecker(ue.GetID(), rules.AsTextLevelChecker(ue.MatchList))
	if ue.IsDefaultOff() {
		lt.MarkDefaultOff(ue.GetID())
	}

	ww := NewCatalanWrongWordInContextRule(nil)
	lt.AddRuleChecker(ww.GetID(), rules.AsSentenceCheckerSimple(ww.Match))
	vr := NewSimpleReplaceVerbsRule(nil)
	lt.AddRuleChecker(vr.GetID(), rules.AsSentenceCheckerSimple(vr.Match))
	bal := NewSimpleReplaceBalearicRule(nil)
	lt.AddRuleChecker(bal.GetID(), rules.AsSentenceCheckerSimple(bal.Match))
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	mw := NewSimpleReplaceMultiwordsRule(nil)
	lt.AddRuleChecker(mw.GetID(), rules.AsSentenceCheckerSimple(mw.Match))
	op := NewReplaceOperationNamesRule(nil)
	lt.AddRuleChecker(op.GetID(), rules.AsSentenceCheckerSimple(op.Match))
	di := NewSimpleReplaceDiacriticsIEC(nil)
	lt.AddRuleChecker(di.GetID(), rules.AsSentenceCheckerSimple(di.Match))
	ang := NewSimpleReplaceAnglicism(nil)
	lt.AddRuleChecker(ang.GetID(), rules.AsSentenceCheckerSimple(ang.Match))

	pf := NewPronomFebleDuplicateRule(nil)
	lt.AddRuleChecker(pf.GetID(), rules.AsSentenceCheckerSimple(pf.Match))
	cc := NewCheckCaseRule(nil)
	lt.AddRuleChecker(cc.GetID(), rules.AsSentenceCheckerSimple(cc.Match))
	adv := NewSimpleReplaceAdverbsMent(nil)
	lt.AddRuleChecker(adv.GetID(), rules.AsSentenceCheckerSimple(adv.Match))

	wrb := NewCatalanWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tres frases successives comencen amb la mateixa paraula.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	// Java CatalanRepeatedWordsRule is commented out — do not invent.

	dnv := NewSimpleReplaceDNVRule(nil)
	lt.AddRuleChecker(dnv.GetID(), rules.AsSentenceCheckerSimple(dnv.Match))
	dnvC := NewSimpleReplaceDNVColloquialRule(nil)
	lt.AddRuleChecker(dnvC.GetID(), rules.AsSentenceCheckerSimple(dnvC.Match))
	dnvS := NewSimpleReplaceDNVSecondaryRule(nil)
	lt.AddRuleChecker(dnvS.GetID(), rules.AsSentenceCheckerSimple(dnvS.Match))

	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	// Java PunctuationMarkAtParagraphEnd(messages, this) → defaultActive true (on).
	ppe := rules.NewPunctuationMarkAtParagraphEnd(map[string]string{
		"punctuation_mark_paragraph_end_msg": "Add a punctuation mark at paragraph end",
	})
	ppe.DefaultOff = false
	lt.AddTextLevelRuleChecker(ppe.GetID(), rules.AsTextLevelChecker(ppe.MatchList))

	remote := NewCatalanRemoteRule()
	lt.AddTextLevelRuleChecker(remote.GetID(), rules.AsTextLevelChecker(remote.MatchList))
	if remote.DefaultOff {
		lt.MarkDefaultOff(remote.GetID())
	}

	split := NewCatalanSplitLongSentenceRule(nil, 60)
	lt.AddTextLevelRuleChecker(split.GetID(), rules.AsTextLevelChecker(split.MatchList))

	ign := NewIgnoreProperNouns()
	lt.AddTextLevelRuleChecker(ign.GetID(), rules.AsTextLevelChecker(ign.MatchList))

	// Java ValencianCatalan.getRelevantRules: super + WordCoherencyValencianRule.
	if isValencian(lt.GetLanguageCode()) {
		wcv := NewWordCoherencyValencianRule(nil)
		lt.AddTextLevelRuleChecker(wcv.GetID(), rules.AsTextLevelChecker(wcv.MatchList))
	}
}
