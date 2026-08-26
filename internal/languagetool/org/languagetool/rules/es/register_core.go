package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCoreSpanishRules ports Spanish.getRelevantRules / createDefaultSpellingRule.
// Java list only — no invent SharedLayout extras (no empty-line / sentence-whitespace /
// whitespace-before-punct / invent TOO_LONG_SENTENCE_ES).
func RegisterCoreSpanishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.PriorityForId = language.SpanishPriorityForId
	languagetool.LanguageAdaptSuggestionByCode["es"] = language.SpanishAdaptSuggestion
	if languagetool.FilterSpanishRuleMatchesHook != nil {
		lt.FilterRuleMatches = languagetool.FilterSpanishRuleMatchesHook
	}
	_ = language.TryWireSpanishVoseoWordTagger()

	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	ub := NewSpanishUnpairedBracketsRule(nil)
	lt.AddTextLevelRuleChecker(ub.GetID(), rules.AsTextLevelChecker(ub.MatchList))

	// Java QuestionMarkRule (text-level match(List<AnalyzedSentence>)).
	qm := NewQuestionMarkRule(nil)
	lt.AddTextLevelRuleChecker(qm.GetID(), rules.AsTextLevelChecker(qm.MatchList))

	sp := NewMorfologikSpanishSpellerRule()
	if p := morfologik.DiscoverLanguageDict(SpanishSpellerDict); p != "" {
		if WireSpanishFilterSpeller(p) {
			sp.IsMisspelled = FilterDictIsMisspelled
		}
	}
	ltRef := lt
	WireSpanishSpellerTagPOS(sp, func(token string) []languagetool.TokenTag {
		if ltRef.TagWord == nil {
			return nil
		}
		return ltRef.TagWord(token)
	})
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

	up := rules.NewUppercaseSentenceStartRule(nil, "es")
	lt.AddRuleChecker(up.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	wr := NewSpanishWordRepeatRule(map[string]string{"repetition": "Repetición de palabra"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	// Java SpanishWikipediaRule — was missing from prior invent registration.
	wiki := NewSpanishWikipediaRule(nil)
	lt.AddRuleChecker(wiki.GetID(), rules.AsSentenceCheckerSimple(wiki.Match))

	ww := NewSpanishWrongWordInContextRule(nil)
	lt.AddRuleChecker(ww.GetID(), rules.AsSentenceCheckerSimple(ww.Match))

	// Java LongSentenceRule(messages, userConfig, 60) → TOO_LONG_SENTENCE.
	ls := rules.NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "This sentence is too long (%d words)",
	}, 60)
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))

	// Java LongParagraphRule(messages, this, userConfig) → maxWords 220, defaultOff.
	lp := rules.NewLongParagraphRule(map[string]string{
		"long_paragraph_rule_msg": "This paragraph is too long (%d words)",
	}, 220)
	lt.AddTextLevelRuleChecker(lp.GetID(), rules.AsTextLevelChecker(lp.MatchList))
	if lp.IsDefaultOff() {
		lt.MarkDefaultOff(lp.GetID())
	}

	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))

	vr := NewSimpleReplaceVerbsRule(nil)
	lt.AddRuleChecker(vr.GetID(), rules.AsSentenceCheckerSimple(vr.Match))

	wrb := NewSpanishWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tres oraciones sucesivas comienzan con la misma palabra.",
		"desc_repetition_beginning_adv":  "Tres oraciones sucesivas comienzan con el mismo adverbio.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))

	rw := NewSpanishRepeatedWordsRule(nil)
	lt.AddTextLevelRuleChecker(rw.GetID(), rules.AsTextLevelChecker(rw.MatchList))
}
